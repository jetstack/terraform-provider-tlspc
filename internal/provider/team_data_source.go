// Copyright (c) Venafi, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"terraform-provider-tlspc/internal/tlspc"
	"terraform-provider-tlspc/internal/validators"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
		MarkdownDescription: "Look up a team by name and return its details.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Team name",
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"role": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: `Role of team, valid options include:
    * SYSTEM_ADMIN
    * PKI_ADMIN
    * PLATFORM_ADMIN
    * RESOURCE_OWNER
    * GUEST`,
			},
			"owners": schema.SetAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of team owner ids",
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validators.Uuid()),
				},
			},
			"members": schema.SetAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of team member ids",
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validators.Uuid()),
				},
			},
			"user_matching_rules": schema.SetNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of rules to add members via SSO claims. Please refer to the [documentation](https://docs.venafi.cloud/vcs-platform/r-team-membership-rule-guidelines/) for detailed rule configuration.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"claim_name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The SSO property that this rule acts on",
						},
						"operator": schema.StringAttribute{
							Required: true,
							MarkdownDescription: `The operator of this rule, valid options:
    * EQUALS
    * NOT_EQUALS
    * CONTAINS
    * NOT_CONTAINS
    * STARTS_WITH
    * ENDS_WITH`,
							Validators: []validator.String{
								stringvalidator.OneOf("EQUALS", "NOT_EQUALS", "CONTAINS", "NOT_CONTAINS", "STARTS_WITH", "ENDS_WITH"),
							},
						},
						"value": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The value to check for",
						},
					},
				},
			},
		},
	}
}

type teamDataSourceModel struct {
	ID                types.String           `tfsdk:"id"`
	Name              types.String           `tfsdk:"name"`
	Role              types.String           `tfsdk:"role"`
	Owners            []types.String         `tfsdk:"owners"`
	Members           []types.String         `tfsdk:"members"`
	UserMatchingRules []teamUserMatchingRule `tfsdk:"user_matching_rules"`
}

type teamUserMatchingRule struct {
	ClaimName types.String `tfsdk:"claim_name"`
	Operator  types.String `tfsdk:"operator"`
	Value     types.String `tfsdk:"value"`
}

// Read refreshes the Terraform state with the latest data.
func (d *teamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state teamDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, err := d.client.GetTeamByName(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving team",
			fmt.Sprintf("Error retrieving team: %s", err.Error()),
		)
		return
	}
	state.ID = types.StringValue(team.ID)
	state.Name = types.StringValue(team.Name)
	state.Role = types.StringValue(team.Role)

	owners := []types.String{}
	for _, v := range team.Owners {
		owners = append(owners, types.StringValue(v))
	}
	state.Owners = owners

	members := []types.String{}
	for _, v := range team.Members {
		members = append(members, types.StringValue(v))
	}
	state.Members = members

	umr := []teamUserMatchingRule{}
	for _, v := range team.UserMatchingRules {
		umr = append(umr, teamUserMatchingRule{
			ClaimName: types.StringValue(v.ClaimName),
			Operator:  types.StringValue(v.Operator),
			Value:     types.StringValue(v.Value),
		})
	}

	if len(umr) > 0 {
		state.UserMatchingRules = umr
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
