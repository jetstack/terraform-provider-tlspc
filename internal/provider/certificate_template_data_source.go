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
	_ datasource.DataSource              = &certTemplateDataSource{}
	_ datasource.DataSourceWithConfigure = &certTemplateDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewCertificateTemplateDataSource() datasource.DataSource {
	return &certTemplateDataSource{}
}

// certTemplateDataSource is the data source implementation.
type certTemplateDataSource struct {
	client *tlspc.Client
}

// Configure adds the provider configured client to the data source.
func (d *certTemplateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *certTemplateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate_template"
}

// Schema defines the schema for the data source.
func (d *certTemplateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up properties of a Certificate Template",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the Certificate Issuing Template",
			},
			"ca_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Type of Certificate Authority (see Certificate Authority Product Option data source)",
			},
			"ca_product_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of a Certificate Authority Product Option",
			},
			"key_reuse": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Allow Private Key Reuse",
			},
		},
	}
}

type certTemplateDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	CAType      types.String `tfsdk:"ca_type"`
	CAProductID types.String `tfsdk:"ca_product_id"`
	KeyReuse    types.Bool   `tfsdk:"key_reuse"`
}

// Read refreshes the Terraform state with the latest data.
func (d *certTemplateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model certTemplateDataSourceModel
	diags := req.Config.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	certTemplates, err := d.client.GetCertTemplates()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving Certificate Templates",
			fmt.Sprintf("Error retrieving Certificate Templates: %s", err.Error()),
		)
		return
	}

	found := false
	for _, v := range certTemplates {
		if model.CAType.ValueString() == v.CertificateAuthorityType && model.Name.ValueString() == v.Name {
			model.ID = types.StringValue(v.ID)
			model.CAProductID = types.StringValue(v.CertificateAuthorityProductOptionID)
			model.KeyReuse = types.BoolValue(v.KeyReuse)
			found = true
			continue
		}
	}

	if !found {
		resp.Diagnostics.AddError(
			"Certificate Template not found",
			"",
		)
		return
	}
	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
}
