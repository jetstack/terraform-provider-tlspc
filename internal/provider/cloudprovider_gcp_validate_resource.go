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
	_ resource.Resource                = &cloudProviderGCPValidateResource{}
	_ resource.ResourceWithConfigure   = &cloudProviderGCPValidateResource{}
	_ resource.ResourceWithImportState = &cloudProviderGCPValidateResource{}
)

type cloudProviderGCPValidateResource struct {
	client *tlspc.Client
}

func NewCloudProviderGCPValidateResource() resource.Resource {
	return &cloudProviderGCPValidateResource{}
}

func (r *cloudProviderGCPValidateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloudprovider_gcp_validate"
}

func (r *cloudProviderGCPValidateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cloudprovider_id": schema.StringAttribute{
				Required: true,
			},
			"validate": schema.BoolAttribute{
				Required: true,
			},
		},
	}
}

func (r *cloudProviderGCPValidateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type cloudProviderGCPValidateResourceModel struct {
	CloudProviderID types.String `tfsdk:"cloudprovider_id"`
	Validate        types.Bool   `tfsdk:"validate"`
}

func (r *cloudProviderGCPValidateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cloudProviderGCPValidateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Validate.ValueBool() {
		resp.Diagnostics.AddError(
			"Error validating GCP Cloud Provider Connection",
			"Validate can only be set to true",
		)
		return
	}

	validated, err := r.client.ValidateCloudProviderGCP(ctx, plan.CloudProviderID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error validating GCP Cloud Provider Connection",
			"Could validate GCP Cloud Provider: "+err.Error(),
		)
		return
	}

	if !validated {
		resp.Diagnostics.AddError(
			"Error validating GCP Cloud Provider Connection",
			"Could validate GCP Cloud Provider connection",
		)
		return
	}

	plan.Validate = types.BoolValue(validated)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *cloudProviderGCPValidateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state cloudProviderGCPValidateResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	validated, err := r.client.GetCloudProviderGCPValidation(ctx, state.CloudProviderID.ValueString())
	// Really, we should parse the error here and conditionally either bomb out, or set the validated status to false
	// The api isn't really built around giving us a good way of determining whether or not it's an error with the request
	// or if the connection requires validation. For now, set the state to false, we can only ever attempt to set it to true,
	// so this should be reasonably safe and sane.
	_ = err
	/*
		if err != nil {
			resp.Diagnostics.AddError(
				"Error retrieving GCP Cloud Provider Connection validation state",
				"Could not retrieve state: "+err.Error(),
			)
			return
		}
	*/

	state.Validate = types.BoolValue(validated)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *cloudProviderGCPValidateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan cloudProviderGCPValidateResourceModel

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

	if !plan.Validate.ValueBool() {
		if state.Validate.ValueBool() {
			resp.Diagnostics.AddError(
				"Error updating GCP Cloud Provider Connection validation",
				"Can not unvalidate connection status",
			)
		} else {
			resp.Diagnostics.AddError(
				"Error validating GCP Cloud Provider Connection",
				"Validate can only be set to true",
			)
		}
		return
	}

	validated, err := r.client.ValidateCloudProviderGCP(ctx, state.CloudProviderID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error validating GCP Cloud Provider Connection",
			"Could validate GCP Cloud Provider: "+err.Error(),
		)
		return
	}

	plan.Validate = types.BoolValue(validated)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *cloudProviderGCPValidateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state cloudProviderGCPValidateResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Can't delete validated state. Nothing to do here.
}

func (r *cloudProviderGCPValidateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
