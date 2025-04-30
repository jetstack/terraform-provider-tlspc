// Copyright (c) Venafi, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"terraform-provider-tlspc/internal/tlspc"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &fireflySubCAResource{}
	_ resource.ResourceWithConfigure   = &fireflySubCAResource{}
	_ resource.ResourceWithImportState = &fireflySubCAResource{}
)

type fireflySubCAResource struct {
	client *tlspc.Client
}

func NewFireflySubCAResource() resource.Resource {
	return &fireflySubCAResource{}
}

func (r *fireflySubCAResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firefly_subca"
}

func (r *fireflySubCAResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				MarkdownDescription: "The name of the Firefly Sub CA Provider",
			},
			"ca_type": schema.StringAttribute{
				Required: true,
			},
			"ca_account_id": schema.StringAttribute{
				Required: true,
			},
			"ca_product_option_id": schema.StringAttribute{
				Required: true,
			},
			"common_name": schema.StringAttribute{
				Required: true,
			},
			"key_algorithm": schema.StringAttribute{
				Required: true,
			},
			"validity_period": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (r *fireflySubCAResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type fireflySubCAResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	CAType            types.String `tfsdk:"ca_type"`
	CAAccountID       types.String `tfsdk:"ca_account_id"`
	CAProductOptionID types.String `tfsdk:"ca_product_option_id"`
	CommonName        types.String `tfsdk:"common_name"`
	KeyAlgorithm      types.String `tfsdk:"key_algorithm"`
	ValidityPeriod    types.String `tfsdk:"validity_period"`
}

func (r *fireflySubCAResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan fireflySubCAResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ff := tlspc.FireflySubCAProvider{
		Name:              plan.Name.ValueString(),
		CAType:            plan.CAType.ValueString(),
		CAAccountID:       plan.CAAccountID.ValueString(),
		CAProductOptionID: plan.CAProductOptionID.ValueString(),
		CommonName:        plan.CommonName.ValueString(),
		KeyAlgorithm:      plan.KeyAlgorithm.ValueString(),
		ValidityPeriod:    plan.ValidityPeriod.ValueString(),
	}
	created, err := r.client.CreateFireflySubCAProvider(ff)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Firefly SubCA Provider",
			"Could not create Firefly SubCA Provider, unexpected error: "+err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(created.ID)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *fireflySubCAResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state fireflySubCAResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ff, err := r.client.GetFireflySubCAProvider(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading FireflyConfig",
			"Could not read FireflyConfig ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(ff.ID)
	state.Name = types.StringValue(ff.Name)
	state.CAType = types.StringValue(ff.CAType)
	state.CAAccountID = types.StringValue(ff.CAAccountID)
	state.CAProductOptionID = types.StringValue(ff.CAProductOptionID)
	state.CommonName = types.StringValue(ff.CommonName)
	state.KeyAlgorithm = types.StringValue(ff.KeyAlgorithm)
	state.ValidityPeriod = types.StringValue(ff.ValidityPeriod)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *fireflySubCAResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state fireflySubCAResourceModel

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

	ff := tlspc.FireflySubCAProvider{
		ID:                state.ID.ValueString(),
		Name:              plan.Name.ValueString(),
		CAType:            plan.CAType.ValueString(),
		CAAccountID:       plan.CAAccountID.ValueString(),
		CAProductOptionID: plan.CAProductOptionID.ValueString(),
		CommonName:        plan.CommonName.ValueString(),
		KeyAlgorithm:      plan.KeyAlgorithm.ValueString(),
		ValidityPeriod:    plan.ValidityPeriod.ValueString(),
	}

	updated, err := r.client.UpdateFireflySubCAProvider(ff)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Firefly SubCA Provider",
			"Could not update Firefly SubCA Provider, unexpected error: "+err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(updated.ID)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *fireflySubCAResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state fireflySubCAResourceModel

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

func (r *fireflySubCAResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
