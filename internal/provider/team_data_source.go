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
	_ datasource.DataSource              = &teamDataSource{}
	_ datasource.DataSourceWithConfigure = &teamDataSource{}
)

// NewTeamDataSource is a helper function to simplify the provider implementation.
func NewTeamDataSource() datasource.DataSource {
	return &teamDataSource{}
}

// teamDataSource is the data source implementation.
type teamDataSource struct {
	client *tlspc.Client
}

// Configure adds the provider configured client to the data source.
func (d *teamDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *teamDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

// Schema defines the schema for the data source.
func (d *teamDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up a team by name and return its ID.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Team name",
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Example response from reading all teams
// {
//   "teams": [
//     {
//       "id": "2341c240-13b0-11f1-a37d-85cf6021b7eb",
//       "name": "Administrators",
//       "systemRoles": [
//         "SYSTEM_ADMIN"
//       ],
//       "productRoles": {},
//       "role": "SYSTEM_ADMIN",
//       "members": [
//         "2c0a67c0-13af-11f1-b2e2-91c6a0bc61cc"
//       ],
//       "owners": [
//         "2c0a67c0-13af-11f1-b2e2-91c6a0bc61cc"
//       ],
//       "companyId": "e10c4d60-dfb1-11ef-9125-8dbc0cd36e8e",
//       "userMatchingRules": [],
//       "modificationDate": "2026-02-27T07:44:25.350+00:00"
//     }
//   ]
// }

type teamDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// type teamResourceModel struct {
// 	ID                types.String       `tfsdk:"id"`
// 	Name              types.String       `tfsdk:"name"`
// 	Role              types.String       `tfsdk:"role"`
// 	Owners            []types.String     `tfsdk:"owners"`
// 	UserMatchingRules []userMatchingRule `tfsdk:"user_matching_rules"`
// }

// type userMatchingRule struct {
// 	ClaimName types.String `tfsdk:"claim_name"`
// 	Operator  types.String `tfsdk:"operator"`
// 	Value     types.String `tfsdk:"value"`
// }

// Read refreshes the Terraform state with the latest data.
func (d *teamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model teamDataSourceModel
	diags := req.Config.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, err := d.client.GetTeamByName(model.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving team",
			fmt.Sprintf("Error retrieving team: %s", err.Error()),
		)
		return
	}
	model.ID = types.StringValue(team.ID)
	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
}
