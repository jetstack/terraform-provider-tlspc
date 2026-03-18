// Copyright (c) Venafi, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"terraform-provider-tlspc/internal/tlspc"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &TenantDataSource{}
	_ datasource.DataSourceWithConfigure = &TenantDataSource{}
)

func NewTenantDataSource() datasource.DataSource {
	return &TenantDataSource{}
}

type TenantDataSource struct {
	client *tlspc.Client
}

// Configure adds the provider configured client to the data source.
func (d *TenantDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*tlspc.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *tlspc.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

// Metadata returns the data source type name.
func (d *TenantDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant"
}

// Schema defines the schema for the data source.
func (d *TenantDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up the ID of a TLS Protect Cloud Tenant based on the auth token used to authenticate to the provider",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the TLS Protect Cloud Tenant",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the TLS Protect Cloud Tenant",
			},
			"url_prefix": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The URL prefix of the TLS Protect Cloud Tenant",
			},
			"domains": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "The domain list associated with the TLS Protect Cloud Tenant",
			},
		},
	}
}

type tenantDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	URLPrefix types.String `tfsdk:"url_prefix"`
	Domains   types.List   `tfsdk:"domains"`
}

// Read refreshes the Terraform state with the latest data.
func (d *TenantDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model tenantDataSourceModel
	diags := req.Config.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	userAccount, err := d.client.GetUserAccounts()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving TLS Protect Cloud Tenant",
			fmt.Sprintf("Error retrieving TLS Protect Cloud Tenant: %s", err.Error()),
		)
		return
	}

	model.ID = types.StringValue(userAccount.Company.ID)
	model.Name = types.StringValue(userAccount.Company.Name)
	model.URLPrefix = types.StringValue(userAccount.Company.URLPrefix)
	model.Domains, diags = types.ListValueFrom(ctx, types.StringType, userAccount.Company.Domains)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
}
