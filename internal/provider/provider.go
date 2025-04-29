// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"terraform-provider-tlspc/internal/tlspc"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &tlspcProvider{}
var _ provider.ProviderWithFunctions = &tlspcProvider{}

// tlspcProvider defines the provider implementation.
type tlspcProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// tlspcProviderModel describes the provider data model.
type tlspcProviderModel struct {
	ApiKey   types.String `tfsdk:"apikey"`
	Endpoint types.String `tfsdk:"endpoint"`
}

func (p *tlspcProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "tlspc"
	resp.Version = p.version
}

func (p *tlspcProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
## Provider for the Venafi TLS Protect Cloud Platform

This provider allows you to manage resources within the Venafi TLS Protect Cloud Platform.
It's at an early stage of development; for production workloads, please ensure that versions are locked and upgrades considered to avoid breaking changes.

### Usage

We recommend that you create a custom user with the [permissions required](https://docs.venafi.cloud/vaas/user-management/about-user-roles/) to manage the necessary resources, and use this user for performing terraform operations.
`,
		Description: "Provider for the Venafi TLS Protect Cloud Platform",
		Attributes: map[string]schema.Attribute{
			"apikey": schema.StringAttribute{
				MarkdownDescription: "API Key",
				Optional:            false,
				Required:            true,
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "TLSPC API Endpoint",
				Optional:            true,
			},
		},
	}
}

func (p *tlspcProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config tlspcProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apikey := os.Getenv("TLSPC_APIKEY")
	endpoint := os.Getenv("TLSPC_ENDPOINT")
	if !config.ApiKey.IsNull() {
		apikey = config.ApiKey.ValueString()
	}
	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}

	client, _ := tlspc.NewClient(apikey, endpoint, p.version)

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *tlspcProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewTeamResource,
		NewServiceAccountResource,
		NewRegistryAccountResource,
		NewPluginResource,
		NewCertificateTemplateResource,
		NewApplicationResource,
	}
}

func (p *tlspcProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewUserDataSource,
		NewCAProductDataSource,
		NewCertificateTemplateDataSource,
	}
}

func (p *tlspcProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &tlspcProvider{
			version: version,
		}
	}
}
