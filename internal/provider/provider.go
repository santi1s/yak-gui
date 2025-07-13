// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/doctolib/yak/internal/helper"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure YakProvider satisfies various provider interfaces.
var _ provider.Provider = &YakProvider{}

// YakProvider defines the provider implementation.
type YakProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// YakProviderModel describes the provider data model.
type YakProviderModel struct {
	Endpoint  types.String `tfsdk:"endpoint"`
	Namespace types.String `tfsdk:"namespace"`
	Region    types.String `tfsdk:"region"`
	VaultRole types.String `tfsdk:"vault_role"`
}

func (p *YakProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "yak"
	resp.Version = p.version
}

func (p *YakProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Vault endpoint",
				Required:            true,
				Optional:            false,
			},
			"namespace": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Vault namespace",
				Optional:            false,
			},
			"region": schema.StringAttribute{
				Required:            false,
				MarkdownDescription: "AWS region to use during Vault authentication",
				Optional:            true,
			},
			"vault_role": schema.StringAttribute{
				Required:            false,
				MarkdownDescription: "Vault role to use during Vault authentication",
				Optional:            true,
			},
		},
	}
}

func (p *YakProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data YakProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := os.Getenv("YAK_TF_PROVIDER_ENDPOINT")
	namespace := os.Getenv("YAK_TF_PROVIDER_NAMESPACE")
	region := os.Getenv("YAK_TF_PROVIDER_REGION")
	vaultRole := os.Getenv("YAK_TF_PROVIDER_VAULT_ROLE")

	if data.Endpoint.ValueString() != "" {
		endpoint = data.Endpoint.ValueString()
	}

	if data.Namespace.ValueString() != "" {
		namespace = data.Namespace.ValueString()
	}

	if data.VaultRole.ValueString() != "" {
		vaultRole = data.VaultRole.ValueString()
	}

	if endpoint == "" {
		resp.Diagnostics.AddError(
			"Missing Endpoint Configuration",
			"While configuring the provider, the endpoint was not found in "+
				"the YAK_TF_PROVIDER_ENDPOINT environment variable or provider "+
				"configuration block endpoint attribute.",
		)
	}

	if namespace == "" {
		resp.Diagnostics.AddError(
			"Missing Namespace Configuration",
			"While configuring the provider, the namespace was not found in "+
				"the YAK_TF_PROVIDER_NAMESPACE environment variable or provider "+
				"configuration block namespace attribute.",
		)
	}

	if vaultRole == "" {
		resp.Diagnostics.AddError(
			"Missing Vault Role Configuration",
			"While configuring the provider, the Vault role was not found in "+
				"the YAK_TF_PROVIDER_VAULT_ROLE environment variable or provider "+
				"configuration block vault_role attribute.",
		)
	}

	// return all errors at once
	if endpoint == "" || namespace == "" || vaultRole == "" {
		return
	}

	if region == "" {
		region = "eu-central-1"
	}

	clients, err := helper.VaultLoginWithAwsAndGetClients(&helper.VaultConfig{
		Endpoints:      []string{endpoint},
		VaultNamespace: namespace,
		AwsProfile:     "",
		AwsRegion:      region,
		VaultRole:      vaultRole,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while initializing Vault Client",
			"An error occured while authenticating with Vault: "+
				err.Error(),
		)
		return
	}

	resp.DataSourceData = clients[0]
	resp.ResourceData = clients[0]
}

func (p *YakProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSecretResource,
	}
}

func (p *YakProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewSecretDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &YakProvider{
			version: version,
		}
	}
}
