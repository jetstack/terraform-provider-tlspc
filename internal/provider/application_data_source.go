// Copyright (c) Venafi, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"terraform-provider-tlspc/internal/tlspc"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ datasource.DataSource              = &userDataSource{}
	_ datasource.DataSourceWithConfigure = &userDataSource{}
)

type applicationDataSource struct {
	client *tlspc.Client
}

func NewApplicationDataSource() datasource.DataSource {
	return &applicationDataSource{}
}

func (r *applicationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (r *applicationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of this resource",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the application",
			},
			"owners": schema.SetAttribute{
				Computed: true,
				ElementType: basetypes.MapType{
					ElemType: types.StringType,
				},
				MarkdownDescription: "A map of owner ids and their type, one of 'USER' or 'TEAM'",
			},
			"ca_template_aliases": schema.MapAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "CA Template alias-to-id mapping for templates available to this application, see example for format",
			},
		},
	}
}

func (r *applicationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	r.client = client
}

type applicationDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Owners            []types.Map  `tfsdk:"owners"`
	CATemplateAliases types.Map    `tfsdk:"ca_template_aliases"`
}

// Read refreshes the Terraform state with the latest data.
func (d *applicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model applicationDataSourceModel
	diags := req.Config.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := d.client.GetApplicationByName(model.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving application",
			fmt.Sprintf("Error retrieving application: %s", err.Error()),
		)
		return
	}

	aliases := map[string]attr.Value{}
	for k, v := range app.CertificateTemplates {
		aliases[k] = types.StringValue(v)
	}

	aliasmap, diags := types.MapValue(types.StringType, aliases)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	owners := []types.Map{}
	for _, v := range app.Owners {
		owner := map[string]attr.Value{
			"type":  types.StringValue(v.Type),
			"owner": types.StringValue(v.ID),
		}
		ownermap, diags := types.MapValue(types.StringType, owner)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		owners = append(owners, ownermap)
	}

	model.ID = types.StringValue(app.ID)
	model.Name = types.StringValue(app.Name)
	model.Owners = owners
	model.CATemplateAliases = aliasmap

	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
}
