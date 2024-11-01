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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &registryAccountResource{}
	_ resource.ResourceWithConfigure   = &registryAccountResource{}
	_ resource.ResourceWithImportState = &registryAccountResource{}
)

type registryAccountResource struct {
	client *tlspc.Client
}

func NewRegistryAccountResource() resource.Resource {
	return &registryAccountResource{}
}

func (r *registryAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_account"
}

func (r *registryAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"owner": schema.StringAttribute{
				Required: true,
			},
			"scopes": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
			},
			"oci_account_name": schema.StringAttribute{
				Computed: true,
			},
			"oci_registry_token": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"credential_lifetime": schema.Int32Attribute{
				Required: true,
			},
		},
	}
}

func (r *registryAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type registryAccountResourceModel struct {
	ID                 types.String   `tfsdk:"id"`
	Name               types.String   `tfsdk:"name"`
	Owner              types.String   `tfsdk:"owner"`
	Scopes             []types.String `tfsdk:"scopes"`
	OciAccountName     types.String   `tfsdk:"oci_account_name"`
	OciRegistryToken   types.String   `tfsdk:"oci_registry_token"`
	CredentialLifetime types.Int32    `tfsdk:"credential_lifetime"`
}

func (r *registryAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan registryAccountResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	scopes := []string{}
	for _, v := range plan.Scopes {
		scopes = append(scopes, v.ValueString())
	}

	registryAccount := tlspc.ServiceAccount{
		Name:               plan.Name.ValueString(),
		Owner:              plan.Owner.ValueString(),
		Scopes:             scopes,
		CredentialLifetime: plan.CredentialLifetime.ValueInt32(),
		AuthenticationType: "ociToken",
	}

	created, err := r.client.CreateServiceAccount(registryAccount)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating registryAccount",
			"Could not create registryAccount, unexpected error: "+err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(created.ID)
	plan.OciAccountName = types.StringValue(created.OciAccountName)
	plan.OciRegistryToken = types.StringValue(created.OciRegistryToken)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *registryAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state registryAccountResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sa, err := r.client.GetServiceAccount(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Registry Account",
			"Could not read registryaccount ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(sa.ID)
	state.Name = types.StringValue(sa.Name)
	state.Owner = types.StringValue(sa.Owner)

	scopes := []types.String{}
	for _, v := range sa.Scopes {
		scopes = append(scopes, types.StringValue(v))
	}
	state.Scopes = scopes

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *registryAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state registryAccountResourceModel

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
	scopes := []string{}
	for _, v := range plan.Scopes {
		scopes = append(scopes, v.ValueString())
	}

	registryAccount := tlspc.ServiceAccount{
		ID:                 state.ID.ValueString(),
		Name:               plan.Name.ValueString(),
		Owner:              plan.Owner.ValueString(),
		Scopes:             scopes,
		CredentialLifetime: plan.CredentialLifetime.ValueInt32(),
		AuthenticationType: "ociToken",
	}

	err := r.client.UpdateServiceAccount(registryAccount)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating registryAccount",
			"Could not update registryAccount, unexpected error: "+err.Error(),
		)
		return
	}
	plan.ID = state.ID
	plan.OciAccountName = state.OciAccountName
	plan.OciRegistryToken = state.OciRegistryToken
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *registryAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state registryAccountResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteServiceAccount(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Service Account",
			"Could not delete Service Account ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *registryAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
