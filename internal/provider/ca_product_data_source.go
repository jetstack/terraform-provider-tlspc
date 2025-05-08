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

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &caProductDataSource{}
	_ datasource.DataSourceWithConfigure = &caProductDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewCAProductDataSource() datasource.DataSource {
	return &caProductDataSource{}
}

// caProductDataSource is the data source implementation.
type caProductDataSource struct {
	client *tlspc.Client
}

// Configure adds the provider configured client to the data source.
func (d *caProductDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *caProductDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ca_product"
}

// Schema defines the schema for the data source.
func (d *caProductDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up the ID of a Certificate Authority Product Option",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"account_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the CA Account",
			},
			"type": schema.StringAttribute{
				Required: true,
				MarkdownDescription: `Type of Certificate Authority, valid values include:
    * BUILTIN
    * DIGICERT
    * GLOBALSIGN
    * ENTRUST
    * MICROSOFT
    * ACME
    * ZTPKI
    * GLOBALSIGNMSSL
    * TPP
    * CONNECTOR`,
			},
			"ca_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of Certificate Authority",
			},
			"product_option": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of Product Option",
			},
		},
	}
}

type caProductDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	AccountID     types.String `tfsdk:"account_id"`
	Type          types.String `tfsdk:"type"`
	CAName        types.String `tfsdk:"ca_name"`
	ProductOption types.String `tfsdk:"product_option"`
}

// Read refreshes the Terraform state with the latest data.
func (d *caProductDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model caProductDataSourceModel
	diags := req.Config.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	caProduct, caAcct, err := d.client.GetCAProductOption(model.Type.ValueString(), model.CAName.ValueString(), model.ProductOption.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving CA Product",
			fmt.Sprintf("Error retrieving CA Product: %s", err.Error()),
		)
		return
	}
	model.ID = types.StringValue(caProduct.ID)
	model.AccountID = types.StringValue(caAcct.ID)
	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
}
