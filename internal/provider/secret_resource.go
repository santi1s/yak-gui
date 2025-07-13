// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/doctolib/yak/internal/cmd/secret"
	"github.com/doctolib/yak/internal/helper"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/vault/api"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &SecretResource{}
var _ resource.ResourceWithImportState = &SecretResource{}

func NewSecretResource() resource.Resource {
	return &SecretResource{}
}

// SecretResource defines the resource implementation.
type SecretResource struct {
	client *api.Client
}

// SecretResourceModel describes the resource data model.
type SecretResourceModel struct {
	Path        types.String    `tfsdk:"path"`
	Version     types.Int64     `tfsdk:"version"`
	DataJSON    jsontypes.Exact `tfsdk:"data_json"`
	Data        types.Map       `tfsdk:"data"`
	Owner       types.String    `tfsdk:"owner"`
	Source      types.String    `tfsdk:"source"`
	Usage       types.String    `tfsdk:"usage"`
	CreatedTime types.String    `tfsdk:"created_time"`
	ID          types.String    `tfsdk:"id"`
}

func (r *SecretResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (r *SecretResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Vault secret resource",

		Attributes: map[string]schema.Attribute{
			"path": schema.StringAttribute{
				MarkdownDescription: "Secret path in Vault (equivalent to `--path` argument on `yak secret create`). Secrets created with this resources are automatically put in a dedicated environment for secrets created by Terraform.",
				Required:            true,
				Optional:            false,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.Int64Attribute{
				Description: "Current version of the secret in Vault.",
				Computed:    true,
			},
			"data_json": schema.StringAttribute{
				Required:    true,
				Computed:    false,
				Sensitive:   true,
				Description: "JSON-encoded secret data to write.",
				CustomType:  jsontypes.ExactType{},
			},
			"data": schema.MapAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "A mapping whose keys are the top-level data keys returned from Vault and whose values are the corresponding values. This map can only represent string data, so any non-string values returned from Vault are serialized as JSON.",
				ElementType: types.StringType,
			},
			"owner": schema.StringAttribute{
				Required:    true,
				Computed:    false,
				Sensitive:   false,
				Description: "Team owner of the secret.",
			},
			"source": schema.StringAttribute{
				Required:    true,
				Computed:    false,
				Sensitive:   false,
				Description: "Source of the secret.",
			},
			"usage": schema.StringAttribute{
				Required:    true,
				Computed:    false,
				Sensitive:   false,
				Description: "Where is the secret used.",
			},
			"created_time": schema.StringAttribute{
				Computed:    true,
				Sensitive:   false,
				Description: "Time at which secret version was created.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the terraform resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *SecretResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *SecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var planData *SecretResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)

	if resp.Diagnostics.HasError() {
		return
	}

	secretPath := secretPrefix + "/" + planData.Path.ValueString()

	// check if the secret exists
	secretVersion, err := secret.GetLatestVersion([]*api.Client{r.client}, secretPath)
	if err != nil {
		resp.Diagnostics.AddError(
			"Secret existence error",
			fmt.Sprintf("Error while checking secret existence before secret creation: %s", err.Error()),
		)
		return
	}
	if secretVersion != -1 {
		resp.Diagnostics.AddError(
			"Secret existence error",
			"Error while creating secret: already exists",
		)
		return
	}

	secretValues := make(map[string]interface{})
	planData.DataJSON.Unmarshal(&secretValues)

	_, err = secret.WriteSecretData([]*api.Client{r.client}, secretPath, secretValues)
	if err != nil {
		resp.Diagnostics.AddError(
			"Secret write error",
			fmt.Sprintf("Error while writing secret data: %s", err.Error()),
		)
		return
	}

	metadata := map[string]interface{}{
		"owner":  planData.Owner.ValueString(),
		"source": planData.Source.ValueString(),
		"usage":  planData.Usage.ValueString(),
	}

	err = secret.WriteSecretMetadata([]*api.Client{r.client}, secretPath, metadata)
	if err != nil {
		resp.Diagnostics.AddError(
			"Secret write error",
			fmt.Sprintf("Error while writing secret metadata: %s", err.Error()),
		)
		return
	}

	values, err := secret.ReadSecretData([]*api.Client{r.client}, secretPath, -1)
	if err != nil {
		resp.Diagnostics.AddError(
			"Secret read error",
			fmt.Sprintf("Error while reading secret after creation: %s", err.Error()),
		)
		return
	}

	if v, ok := values.Data["data"].(map[string]interface{}); ok {
		planData.Data, _ = types.MapValueFrom(ctx, types.StringType, helper.SerializeVaultDataMapToString(v))
	} else {
		resp.Diagnostics.AddError(
			"Secret read error",
			"Error while reading secret data after creation. Data is not a map[string]interface{}.",
		)
		return
	}

	if metadata, ok := values.Data["metadata"].(map[string]interface{}); ok {
		if customMetadata, ok := metadata["custom_metadata"].(map[string]interface{}); ok {
			planData.Owner = types.StringValue(customMetadata["owner"].(string))
			planData.Source = types.StringValue(customMetadata["source"].(string))
			planData.Usage = types.StringValue(customMetadata["usage"].(string))
		} else {
			resp.Diagnostics.AddError(
				"Secret read error",
				"Error while reading secret custom_metadata after creation. custom_metadata is not a map[string]interface{}.",
			)
			return
		}

		if createdTime, ok := metadata["created_time"].(string); ok {
			planData.CreatedTime = types.StringValue(createdTime)
		} else {
			resp.Diagnostics.AddError(
				"Secret read error",
				"Error while reading secret created_time after creation. created_time is not a string.",
			)
			return
		}

		if version, ok := metadata["version"].(json.Number); ok {
			ver, err := strconv.Atoi(version.String())
			if err != nil {
				resp.Diagnostics.AddError(
					"Secret read error",
					"Error while reading converting secret version from string to integer.",
				)
				return
			}
			planData.Version = types.Int64Value(int64(ver))
		} else {
			resp.Diagnostics.AddError(
				"Secret read error",
				"Error while reading secret version after creation. version is not a json.Number.",
			)
			return
		}
	} else {
		resp.Diagnostics.AddError(
			"Secret read error",
			"Error while reading secret metadata after creation. metadata is not a map[string]interface{}.",
		)
	}

	planData.ID = types.StringValue(planData.Path.ValueString())

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, fmt.Sprintf("created secret version %d of secret %s", planData.Version.ValueInt64(), planData.Path.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &planData)...)
}

func (r *SecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var planData *SecretResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	secretPath := secretPrefix + "/" + planData.Path.ValueString()
	values, err := secret.ReadSecretData([]*api.Client{r.client}, secretPath, -1)
	if err != nil {
		resp.Diagnostics.AddError(
			"Secret read error",
			fmt.Sprintf("Error while reading secret after creation: %s", err.Error()),
		)
		return
	}

	if v, ok := values.Data["data"].(map[string]interface{}); ok {
		planData.Data, _ = types.MapValueFrom(ctx, types.StringType, helper.SerializeVaultDataMapToString(v))

		jsonString, _ := json.Marshal(helper.SerializeVaultDataMapToString(v))
		planData.DataJSON = jsontypes.NewExactValue(string(jsonString))
	} else {
		resp.Diagnostics.AddError(
			"Secret read error",
			"Error while reading secret data after creation. Data is not a map[string]interface{}.",
		)
		return
	}

	if metadata, ok := values.Data["metadata"].(map[string]interface{}); ok {
		if customMetadata, ok := metadata["custom_metadata"].(map[string]interface{}); ok {
			planData.Owner = types.StringValue(customMetadata["owner"].(string))
			planData.Source = types.StringValue(customMetadata["source"].(string))
			planData.Usage = types.StringValue(customMetadata["usage"].(string))
		} else {
			resp.Diagnostics.AddError(
				"Secret read error",
				"Error while reading secret custom_metadata after creation. custom_metadata is not a map[string]interface{}.",
			)
			return
		}

		if createdTime, ok := metadata["created_time"].(string); ok {
			planData.CreatedTime = types.StringValue(createdTime)
		} else {
			resp.Diagnostics.AddError(
				"Secret read error",
				"Error while reading secret created_time after creation. created_time is not a string.",
			)
			return
		}

		if version, ok := metadata["version"].(json.Number); ok {
			ver, err := strconv.Atoi(version.String())
			if err != nil {
				resp.Diagnostics.AddError(
					"Secret read error",
					"Error while reading converting secret version from string to integer.",
				)
				return
			}
			planData.Version = types.Int64Value(int64(ver))
		} else {
			resp.Diagnostics.AddError(
				"Secret read error",
				"Error while reading secret version after creation. version is not a json.Number.",
			)
			return
		}
	} else {
		resp.Diagnostics.AddError(
			"Secret read error",
			"Error while reading secret metadata after creation. metadata is not a map[string]interface{}.",
		)
	}

	planData.ID = types.StringValue(planData.Path.ValueString())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &planData)...)
}

func (r *SecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planData *SecretResourceModel
	var stateData *SecretResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	req.State.Get(ctx, &stateData)

	if resp.Diagnostics.HasError() {
		return
	}

	secretValues := make(map[string]interface{})
	planData.DataJSON.Unmarshal(&secretValues)
	stateSecretValues := make(map[string]interface{})
	stateData.DataJSON.Unmarshal(&stateSecretValues)

	newSecretValues := make(map[string]interface{})
	for k, v := range secretValues {
		newSecretValues[k] = v
	}
	for k := range stateSecretValues {
		if _, ok := secretValues[k]; !ok {
			newSecretValues[k] = nil
		}
	}

	secretPath := secretPrefix + "/" + planData.Path.ValueString()

	jsonDataJSONStringFromPlan, _ := json.Marshal(planData.DataJSON.ValueString())
	jsonDataJSONStringFromState, _ := json.Marshal(stateData.DataJSON.ValueString())
	if string(jsonDataJSONStringFromPlan) != string(jsonDataJSONStringFromState) {
		_, err := secret.PatchSecretData([]*api.Client{r.client}, secretPath, newSecretValues)
		if err != nil {
			resp.Diagnostics.AddError(
				"Secret write error",
				fmt.Sprintf("Error while updating secret: %s", err.Error()),
			)
			return
		}
	}

	if planData.Owner.ValueString() != stateData.Owner.ValueString() || planData.Source.ValueString() != stateData.Source.ValueString() || planData.Usage.ValueString() != stateData.Usage.ValueString() {
		payload := map[string]interface{}{
			"custom_metadata": map[string]interface{}{
				"owner":  planData.Owner.ValueString(),
				"source": planData.Source.ValueString(),
				"usage":  planData.Usage.ValueString(),
			},
		}

		_, err := secret.PatchSecretMetadata([]*api.Client{r.client}, secretPath, payload)
		if err != nil {
			resp.Diagnostics.AddError(
				"Secret write error",
				fmt.Sprintf("Error while updating secret metadata: %s", err.Error()),
			)
			return
		}
	}

	values, err := secret.ReadSecretData([]*api.Client{r.client}, secretPath, -1)
	if err != nil {
		resp.Diagnostics.AddError(
			"Secret read error",
			fmt.Sprintf("Error while reading secret after creation: %s", err.Error()),
		)
		return
	}

	if v, ok := values.Data["data"].(map[string]interface{}); ok {
		planData.Data, _ = types.MapValueFrom(ctx, types.StringType, helper.SerializeVaultDataMapToString(v))
	} else {
		resp.Diagnostics.AddError(
			"Secret read error",
			"Error while reading secret data after creation. Data is not a map[string]interface{}.",
		)
		return
	}

	if metadata, ok := values.Data["metadata"].(map[string]interface{}); ok {
		if customMetadata, ok := metadata["custom_metadata"].(map[string]interface{}); ok {
			planData.Owner = types.StringValue(customMetadata["owner"].(string))
			planData.Source = types.StringValue(customMetadata["source"].(string))
			planData.Usage = types.StringValue(customMetadata["usage"].(string))
		} else {
			resp.Diagnostics.AddError(
				"Secret read error",
				"Error while reading secret custom_metadata after creation. custom_metadata is not a map[string]interface{}.",
			)
			return
		}

		if createdTime, ok := metadata["created_time"].(string); ok {
			planData.CreatedTime = types.StringValue(createdTime)
		} else {
			resp.Diagnostics.AddError(
				"Secret read error",
				"Error while reading secret created_time after creation. created_time is not a string.",
			)
			return
		}

		if version, ok := metadata["version"].(json.Number); ok {
			ver, err := strconv.Atoi(version.String())
			if err != nil {
				resp.Diagnostics.AddError(
					"Secret read error",
					"Error while reading converting secret version from string to integer.",
				)
				return
			}
			planData.Version = types.Int64Value(int64(ver))
		} else {
			resp.Diagnostics.AddError(
				"Secret read error",
				"Error while reading secret version after creation. version is not a json.Number.",
			)
			return
		}
	} else {
		resp.Diagnostics.AddError(
			"Secret read error",
			"Error while reading secret metadata after creation. metadata is not a map[string]interface{}.",
		)
	}

	tflog.Trace(ctx, fmt.Sprintf("created secret version %d of secret %s", planData.Version.ValueInt64(), planData.Path.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &planData)...)
}

func (r *SecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *SecretResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	secretPath := secretPrefix + "/" + data.Path.ValueString()

	_, err := r.client.Logical().Delete("kv/metadata/" + secretPath)
	if err != nil {
		resp.Diagnostics.AddError(
			"Secret delete error",
			fmt.Sprintf("Error while deleting secret: %s", err.Error()),
		)
		return
	}

	_, err = r.client.Logical().Delete("kv/metadata/ci/" + secretPath)
	if err != nil {
		resp.Diagnostics.AddError(
			"Secret delete error",
			fmt.Sprintf("Error while deleting secret: %s", err.Error()),
		)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted secret %s", data.Path.String()))
}

func (r *SecretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("path"), req, resp)
}
