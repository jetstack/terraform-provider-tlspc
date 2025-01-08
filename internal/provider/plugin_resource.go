// Copyright (c) Venafi, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"terraform-provider-tlspc/internal/tlspc"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &pluginResource{}
	_ resource.ResourceWithConfigure   = &pluginResource{}
	_ resource.ResourceWithImportState = &pluginResource{}
)

type pluginResource struct {
	client *tlspc.Client
}

func NewPluginResource() resource.Resource {
	return &pluginResource{}
}

func (r *pluginResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_plugin"
}

func (r *pluginResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Required: true,
			},
			"manifest": schema.StringAttribute{
				Required:   true,
				CustomType: jsontypes.NormalizedType{},
			},
		},
	}
}

func (r *pluginResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type pluginResourceModel struct {
	ID       types.String         `tfsdk:"id"`
	Type     types.String         `tfsdk:"type"`
	Manifest jsontypes.Normalized `tfsdk:"manifest"`
}

func (r *pluginResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan pluginResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var manifest any
	err := json.Unmarshal([]byte(plan.Manifest.ValueString()), &manifest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating plugin",
			"Could not create team, invalid manifest: "+err.Error(),
		)
		return
	}

	plugin := tlspc.Plugin{
		ID:       plan.ID.ValueString(),
		Type:     plan.Type.ValueString(),
		Manifest: manifest,
	}

	created, err := r.client.CreatePlugin(plugin)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating plugin",
			"Could not create plugin, unexpected error: "+err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(created.ID)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *pluginResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state pluginResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plugin, err := r.client.GetPlugin(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Plugin",
			"Could not read plugin ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(plugin.ID)
	state.Type = types.StringValue(plugin.Type)
	stateManifest, err := json.Marshal(plugin.Manifest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Plugin",
			"Could not read plugin manifest: "+err.Error(),
		)
		return
	}
	state.Manifest = jsontypes.NewNormalizedValue(string(stateManifest))

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *pluginResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan pluginResourceModel

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

	var manifest any
	err := json.Unmarshal([]byte(plan.Manifest.ValueString()), &manifest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Plugin",
			"Could not read plugin manifest: "+err.Error(),
		)
		return
	}
	plugin := tlspc.Plugin{
		ID:       state.ID.ValueString(),
		Type:     plan.Type.ValueString(),
		Manifest: manifest,
	}
	err = r.client.UpdatePlugin(plugin)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Plugin",
			"Could not update plugin: "+err.Error(),
		)
		return
	}

	plan.ID = state.ID
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *pluginResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state pluginResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePlugin(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Plugin",
			"Could not delete plugin ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *pluginResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
