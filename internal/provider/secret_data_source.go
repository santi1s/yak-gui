// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/doctolib/yak/internal/cmd/secret"
	"github.com/doctolib/yak/internal/helper"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/vault/api"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &SecretDataSource{}

func NewSecretDataSource() datasource.DataSource {
	return &SecretDataSource{}
}

// SecretFileDataSource defines the data source implementation.
type SecretDataSource struct {
	client *api.Client
}

// SecretDataSourceModel describes the data source data model.
type SecretDataSourceModel struct {
	Path           types.String `tfsdk:"path"`
	Version        types.Int64  `tfsdk:"version"`
	Data           types.Map    `tfsdk:"data"`
	CustomMetadata types.Map    `tfsdk:"custom_metadata"`
	CreatedTime    types.String `tfsdk:"created_time"`
	DeletionTime   types.String `tfsdk:"deletion_time"`
	Destroyed      types.Bool   `tfsdk:"destroyed"`
	ID             types.String `tfsdk:"id"`
}

func (d *SecretDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (d *SecretDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Vault secret data source for secrets created by Terraform. This data source is not suitable for secrets created by other means.",

		Attributes: map[string]schema.Attribute{
			"path": schema.StringAttribute{
				MarkdownDescription: "Secret path in Vault (equivalent to `--path` argument on `yak secret create`). Secrets read with this resources are automatically get from a dedicated environment for secrets created by Terraform.",
				Required:            true,
				Optional:            false,
			},
			"version": schema.Int64Attribute{
				Description: "Current version of the secret in Vault.",
				Required:    true,
				Optional:    false,
			},
			"data": schema.MapAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "A mapping whose keys are the top-level data keys returned from Vault and whose values are the corresponding values. This map can only represent string data, so any non-string values returned from Vault are serialized as JSON.",
				ElementType: types.StringType,
			},
			"custom_metadata": schema.MapAttribute{
				Computed:            true,
				Sensitive:           false,
				ElementType:         types.StringType,
				MarkdownDescription: "Custom metadata of the secret (including `owner`, `source` and `usage`).",
			},
			"created_time": schema.StringAttribute{
				Computed:    true,
				Sensitive:   false,
				Description: "Time at which secret version was created.",
			},
			"deletion_time": schema.StringAttribute{
				Computed:    true,
				Sensitive:   false,
				Description: "Deletion time for the secret version.",
			},
			"destroyed": schema.BoolAttribute{
				Computed:    true,
				Sensitive:   false,
				Description: "Indicates whether the secret version has been destroyed.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the terraform resource.",
			},
		},
	}
}

func (d *SecretDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *SecretDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SecretDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	values, err := secret.ReadSecretData([]*api.Client{d.client}, secretPrefix+"/"+data.Path.ValueString(), int(data.Version.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Secret read Error",
			fmt.Sprintf("Error while reading secret: %s", err.Error()),
		)
		return
	}

	if v, ok := values.Data["data"].(map[string]interface{}); ok {
		data.Data, _ = types.MapValueFrom(ctx, types.StringType, helper.SerializeVaultDataMapToString(v))
	} else {
		resp.Diagnostics.AddError(
			"Secret read error",
			"Error while reading secret data after creation. Data is not a map[string]interface{}.",
		)
		return
	}

	if metadata, ok := values.Data["metadata"].(map[string]interface{}); ok {
		if customMetadata, ok := metadata["custom_metadata"].(map[string]interface{}); ok {
			data.CustomMetadata, _ = types.MapValueFrom(ctx, types.StringType, helper.SerializeVaultDataMapToString(customMetadata))
		} else {
			resp.Diagnostics.AddError(
				"Secret read error",
				"Error while reading secret custom_metadata after creation. custom_metadata is not a map[string]interface{}.",
			)
			return
		}

		if createdTime, ok := metadata["created_time"].(string); ok {
			data.CreatedTime = types.StringValue(createdTime)
		} else {
			resp.Diagnostics.AddError(
				"Secret read error",
				"Error while reading secret created_time after creation. created_time is not a string.",
			)
			return
		}

		if deletionTime, ok := metadata["deletion_time"].(string); ok {
			data.DeletionTime = types.StringValue(deletionTime)
		} else {
			resp.Diagnostics.AddError(
				"Data Source Read Error",
				"Error while reading deletion_time value",
			)
			return
		}

		if destroyed, ok := metadata["destroyed"].(bool); ok {
			data.Destroyed = types.BoolValue(destroyed)
		} else {
			resp.Diagnostics.AddError(
				"Data Source Read Error",
				"Error while reading destroyed value",
			)
			return
		}
	} else {
		resp.Diagnostics.AddError(
			"Data Source Read Error",
			"Error while reading metadata value",
		)
	}

	data.ID = types.StringValue(data.Path.ValueString() + "-" + data.Version.String())

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, fmt.Sprintf("read secret %s in version %d", data.Path.ValueString(), data.Version.ValueInt64()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
