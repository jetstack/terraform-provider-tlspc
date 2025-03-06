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
	_ resource.Resource                = &cloudProviderGCPResource{}
	_ resource.ResourceWithConfigure   = &cloudProviderGCPResource{}
	_ resource.ResourceWithImportState = &cloudProviderGCPResource{}
)

type cloudProviderGCPResource struct {
	client *tlspc.Client
}

func NewCloudProviderGCPResource() resource.Resource {
	return &cloudProviderGCPResource{}
}

func (r *cloudProviderGCPResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloudprovider_gcp"
}

func (r *cloudProviderGCPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"issuer_url": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"team": schema.StringAttribute{
				Required: true,
			},
			"service_account_email": schema.StringAttribute{
				Required: true,
			},
			"project_number": schema.Int64Attribute{
				Required: true,
			},
			"workload_identity_pool_id": schema.StringAttribute{
				Required: true,
			},
			"workload_identity_pool_provider_id": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (r *cloudProviderGCPResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type cloudProviderGCPResourceModel struct {
	ID                             types.String `tfsdk:"id"`
	IssuerUrl                      types.String `tfsdk:"issuer_url"`
	Name                           types.String `tfsdk:"name"`
	Team                           types.String `tfsdk:"team"`
	ServiceAccountEmail            types.String `tfsdk:"service_account_email"`
	ProjectNumber                  types.Int64  `tfsdk:"project_number"`
	WorkloadIdentityPoolId         types.String `tfsdk:"workload_identity_pool_id"`
	WorkloadIdentityPoolProviderId types.String `tfsdk:"workload_identity_pool_provider_id"`
}

func (r *cloudProviderGCPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cloudProviderGCPResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := tlspc.CloudProviderGCP{
		Name:                           plan.Name.ValueString(),
		Team:                           plan.Team.ValueString(),
		ServiceAccountEmail:            plan.ServiceAccountEmail.ValueString(),
		ProjectNumber:                  plan.ProjectNumber.ValueInt64(),
		WorkloadIdentityPoolId:         plan.WorkloadIdentityPoolId.ValueString(),
		WorkloadIdentityPoolProviderId: plan.WorkloadIdentityPoolProviderId.ValueString(),
	}

	created, err := r.client.CreateCloudProviderGCP(ctx, p)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating GCP Cloud Provider",
			"Could not create GCP Cloud Provider: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(created.ID)
	plan.IssuerUrl = types.StringValue(created.IssuerUrl)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *cloudProviderGCPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state cloudProviderGCPResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cp, err := r.client.GetCloudProviderGCP(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving GCP Cloud Provider",
			"Could not find GCP Cloud Provider: "+err.Error(),
		)
		return
	}

	state.IssuerUrl = types.StringValue(cp.IssuerUrl)
	state.Name = types.StringValue(cp.Name)
	state.Team = types.StringValue(cp.Team)
	state.ServiceAccountEmail = types.StringValue(cp.ServiceAccountEmail)
	state.ProjectNumber = types.Int64Value(cp.ProjectNumber)
	state.WorkloadIdentityPoolId = types.StringValue(cp.WorkloadIdentityPoolId)
	state.WorkloadIdentityPoolProviderId = types.StringValue(cp.WorkloadIdentityPoolProviderId)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *cloudProviderGCPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan cloudProviderGCPResourceModel

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

	cp := tlspc.CloudProviderGCP{
		ID:                             state.ID.ValueString(),
		Name:                           plan.Name.ValueString(),
		Team:                           plan.Team.ValueString(),
		ServiceAccountEmail:            plan.ServiceAccountEmail.ValueString(),
		ProjectNumber:                  plan.ProjectNumber.ValueInt64(),
		WorkloadIdentityPoolId:         plan.WorkloadIdentityPoolId.ValueString(),
		WorkloadIdentityPoolProviderId: plan.WorkloadIdentityPoolProviderId.ValueString(),
	}

	updated, err := r.client.UpdateCloudProviderGCP(ctx, cp)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating GCP Cloud Provider",
			"Could not update GCP Cloud Provider, unexpected error: "+err.Error(),
		)
		return
	}
	plan.IssuerUrl = types.StringValue(updated.IssuerUrl)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *cloudProviderGCPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state cloudProviderGCPResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteCloudProviderGCP(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating GCP Cloud Provider",
			"Could not updating GCP Cloud Provider: "+err.Error(),
		)
		return
	}
}

func (r *cloudProviderGCPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
