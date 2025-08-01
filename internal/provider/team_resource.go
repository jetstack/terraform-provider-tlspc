// Copyright (c) Venafi, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"reflect"

	"terraform-provider-tlspc/internal/tlspc"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &teamResource{}
	_ resource.ResourceWithConfigure   = &teamResource{}
	_ resource.ResourceWithImportState = &teamResource{}
)

type teamResource struct {
	client *tlspc.Client
}

func NewTeamResource() resource.Resource {
	return &teamResource{}
}

func (r *teamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (r *teamResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name",
			},
			"role": schema.StringAttribute{
				Required: true,
				MarkdownDescription: `Role of team, valid options include:
    * SYSTEM_ADMIN
    * PKI_ADMIN
    * PLATFORM_ADMIN
    * RESOURCE_OWNER
    * GUEST`,
			},
			"owners": schema.SetAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of user ids",
			},
			"user_matching_rules": schema.SetNestedAttribute{
				Optional:            true,
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

func (r *teamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type teamResourceModel struct {
	ID                types.String       `tfsdk:"id"`
	Name              types.String       `tfsdk:"name"`
	Role              types.String       `tfsdk:"role"`
	Owners            []types.String     `tfsdk:"owners"`
	UserMatchingRules []userMatchingRule `tfsdk:"user_matching_rules"`
}

type userMatchingRule struct {
	ClaimName types.String `tfsdk:"claim_name"`
	Operator  types.String `tfsdk:"operator"`
	Value     types.String `tfsdk:"value"`
}

func (r *teamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan teamResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	owners := []string{}
	for _, v := range plan.Owners {
		owners = append(owners, v.ValueString())
	}

	umr := []tlspc.UserMatchingRule{}
	for _, v := range plan.UserMatchingRules {
		umr = append(umr, tlspc.UserMatchingRule{
			ClaimName: v.ClaimName.ValueString(),
			Operator:  v.Operator.ValueString(),
			Value:     v.Value.ValueString(),
		})
	}

	team := tlspc.Team{
		Name:              plan.Name.ValueString(),
		Role:              plan.Role.ValueString(),
		Owners:            owners,
		Members:           []string{},
		UserMatchingRules: umr,
	}

	created, err := r.client.CreateTeam(team)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating team",
			"Could not create team, unexpected error: "+err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(created.ID)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *teamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state teamResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, err := r.client.GetTeam(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Team",
			"Could not read team ID "+state.ID.ValueString()+": "+err.Error(),
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

	umr := []userMatchingRule{}
	for _, v := range team.UserMatchingRules {
		umr = append(umr, userMatchingRule{
			ClaimName: types.StringValue(v.ClaimName),
			Operator:  types.StringValue(v.Operator),
			Value:     types.StringValue(v.Value),
		})
	}

	state.UserMatchingRules = umr

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *teamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan teamResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Name != plan.Name || state.Role != plan.Role || !reflect.DeepEqual(state.UserMatchingRules, plan.UserMatchingRules) {
		umr := []tlspc.UserMatchingRule{}
		for _, v := range plan.UserMatchingRules {
			umr = append(umr, tlspc.UserMatchingRule{
				ClaimName: v.ClaimName.ValueString(),
				Operator:  v.Operator.ValueString(),
				Value:     v.Value.ValueString(),
			})
		}
		team := tlspc.Team{
			ID:                state.ID.ValueString(),
			Name:              plan.Name.ValueString(),
			Role:              plan.Role.ValueString(),
			UserMatchingRules: umr,
		}
		_, err := r.client.UpdateTeam(team)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Team",
				"Could not update team ID "+state.ID.ValueString()+": "+err.Error(),
			)
			return
		}
	}
	stateOwners := map[string]bool{}
	planOwners := map[string]bool{}
	for _, v := range state.Owners {
		stateOwners[v.ValueString()] = true
	}
	for _, v := range plan.Owners {
		planOwners[v.ValueString()] = true
	}
	addOwners := []string{}
	removeOwners := []string{}
	for k := range stateOwners {
		if _, exists := planOwners[k]; !exists {
			removeOwners = append(removeOwners, k)
		}
	}
	for k := range planOwners {
		if _, exists := stateOwners[k]; !exists {
			addOwners = append(addOwners, k)
		}
	}
	if len(addOwners) > 0 {
		_, err := r.client.AddTeamOwners(state.ID.ValueString(), addOwners)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Team",
				"Could not update team ID "+state.ID.ValueString()+": "+err.Error(),
			)
			return
		}
	}
	if len(removeOwners) > 0 {
		_, err := r.client.RemoveTeamOwners(state.ID.ValueString(), removeOwners)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Team",
				"Could not update team ID "+state.ID.ValueString()+": "+err.Error(),
			)
			return
		}
	}

	plan.ID = state.ID
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *teamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state teamResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteTeam(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Team",
			"Could not delete team ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *teamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
