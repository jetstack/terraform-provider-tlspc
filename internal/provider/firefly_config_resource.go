// Copyright (c) Venafi, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"terraform-provider-tlspc/internal/tlspc"
	"terraform-provider-tlspc/internal/validators"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &fireflyConfigResource{}
	_ resource.ResourceWithConfigure   = &fireflyConfigResource{}
	_ resource.ResourceWithImportState = &fireflyConfigResource{}
)

type fireflyConfigResource struct {
	client *tlspc.Client
}

func NewFireflyConfigResource() resource.Resource {
	return &fireflyConfigResource{}
}

func (r *fireflyConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firefly_config"
}

func (r *fireflyConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "The ID of this resource",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the Firefly Configuration",
			},
			"subca_provider": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the Firefly SubCA Provider",
				Validators: []validator.String{
					validators.Uuid(),
				},
			},
			"service_accounts": schema.SetAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "A list of service account IDs",
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validators.Uuid()),
				},
			},
			"policies": schema.SetAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "A list of Firefly Issuance Policy IDs",
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validators.Uuid()),
				},
			},
		},
	}
}

func (r *fireflyConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type fireflyConfigResourceModel struct {
	ID              types.String   `tfsdk:"id"`
	Name            types.String   `tfsdk:"name"`
	SubCAProvider   types.String   `tfsdk:"subca_provider"`
	ServiceAccounts []types.String `tfsdk:"service_accounts"`
	Policies        []types.String `tfsdk:"policies"`
}

func (r *fireflyConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan fireflyConfigResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sa := []string{}
	for _, v := range plan.ServiceAccounts {
		sa = append(sa, v.ValueString())
	}

	policies := []string{}
	for _, v := range plan.Policies {
		policies = append(policies, v.ValueString())
	}

	ff := tlspc.FireflyConfig{
		Name:              plan.Name.ValueString(),
		SubCAProviderId:   plan.SubCAProvider.ValueString(),
		PolicyIds:         policies,
		ServiceAccountIds: sa,
		MinTLSVersion:     "TLS13",
		//ClientAuthentication: tlspc.ClientAuthentication{},
	}
	created, err := r.client.CreateFireflyConfig(ff)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating FireflyConfig",
			"Could not create FireflyConfig, unexpected error: "+err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(created.ID)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *fireflyConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state fireflyConfigResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ff, err := r.client.GetFireflyConfig(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading FireflyConfig",
			"Could not read FireflyConfig ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(ff.ID)
	state.Name = types.StringValue(ff.Name)

	sa := []types.String{}
	for _, v := range ff.ServiceAccountIds {
		sa = append(sa, types.StringValue(v))
	}
	state.ServiceAccounts = sa

	policies := []types.String{}
	for _, v := range ff.Policies {
		policies = append(policies, types.StringValue(v.ID))
	}
	state.Policies = policies

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *fireflyConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state fireflyConfigResourceModel

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
	sa := []string{}
	for _, v := range plan.ServiceAccounts {
		sa = append(sa, v.ValueString())
	}

	policies := []string{}
	for _, v := range plan.Policies {
		policies = append(policies, v.ValueString())
	}

	ff := tlspc.FireflyConfig{
		ID:                state.ID.ValueString(),
		Name:              plan.Name.ValueString(),
		SubCAProviderId:   plan.SubCAProvider.ValueString(),
		PolicyIds:         policies,
		ServiceAccountIds: sa,
		MinTLSVersion:     "TLS13",
		/*
			ClientAuthentication: tlspc.ClientAuthentication{
				Type: "None",
			},
		*/
	}

	updated, err := r.client.UpdateFireflyConfig(ff)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating FireflyConfig",
			"Could not update FireflyConfig, unexpected error: "+err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(updated.ID)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *fireflyConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state fireflyConfigResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteFireflyConfig(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting FireflyConfig",
			"Could not delete FireflyConfig ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *fireflyConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
