// Copyright (c) Venafi, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"terraform-provider-tlspc/internal/tlspc"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                = &applicationResource{}
	_ resource.ResourceWithConfigure   = &applicationResource{}
	_ resource.ResourceWithImportState = &applicationResource{}
)

type applicationResource struct {
	client *tlspc.Client
}

func NewApplicationResource() resource.Resource {
	return &applicationResource{}
}

func (r *applicationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (r *applicationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				MarkdownDescription: "The name of the application",
			},
			"owners": schema.SetAttribute{
				Required: true,
				ElementType: basetypes.MapType{
					ElemType: types.StringType,
				},
				MarkdownDescription: "A map of owner ids, see example for format",
			},
			"ca_template_aliases": schema.MapAttribute{
				// Not an API required field. In order to update with blank map for application deletion, this must be optional.
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "CA Template alias-to-id mapping for templates available to this application, see example for format",
			},
		},
	}
}

func (r *applicationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type applicationResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Owners            []types.Map  `tfsdk:"owners"`
	CATemplateAliases types.Map    `tfsdk:"ca_template_aliases"`
}

func (r *applicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan applicationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	owners := []tlspc.OwnerAndType{}
	for _, v := range plan.Owners {
		m := v.Elements()
		// TODO: Work out how you're supposed to get an unquoted string out
		kind := strings.Trim(m["type"].String(), `"`)
		ownerId := strings.Trim(m["owner"].String(), `"`)
		if kind != "USER" && kind != "TEAM" {
			resp.Diagnostics.AddError(
				"Error creating application",
				"Could not create application, unsupported owner type: "+kind,
			)
			return
		}
		if ownerId == "" {
			resp.Diagnostics.AddError(
				"Error creating application",
				"Could not create application, undefined owner",
			)
			return
		}
		owner := tlspc.OwnerAndType{
			ID:   ownerId,
			Type: kind,
		}
		owners = append(owners, owner)
	}

	aliases := map[string]string{}
	for k, v := range plan.CATemplateAliases.Elements() {
		aliases[k] = strings.Trim(v.String(), `"`)
	}

	application := tlspc.Application{
		Name:                 plan.Name.ValueString(),
		Owners:               owners,
		CertificateTemplates: aliases,
	}
	created, err := r.client.CreateApplication(application)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating application",
			"Could not create application, unexpected error: "+err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(created.ID)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *applicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state applicationResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := r.client.GetApplication(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Application",
			"Could not read application ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(app.ID)
	state.Name = types.StringValue(app.Name)

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
	state.Owners = owners

	aliases := map[string]attr.Value{}
	for k, v := range app.CertificateTemplates {
		aliases[k] = types.StringValue(v)
	}

	aliasmap, diags := types.MapValue(types.StringType, aliases)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.CATemplateAliases = aliasmap

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *applicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state applicationResourceModel

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
	owners := []tlspc.OwnerAndType{}
	for _, v := range plan.Owners {
		m := v.Elements()
		// TODO: Work out how you're supposed to get an unquoted string out
		kind := strings.Trim(m["type"].String(), `"`)
		ownerId := strings.Trim(m["owner"].String(), `"`)
		if kind != "USER" && kind != "TEAM" {
			resp.Diagnostics.AddError(
				"Error creating application",
				"Could not create application, unsupported owner type: "+kind,
			)
			return
		}
		if ownerId == "" {
			resp.Diagnostics.AddError(
				"Error creating application",
				"Could not create application, undefined owner",
			)
			return
		}
		owner := tlspc.OwnerAndType{
			ID:   ownerId,
			Type: kind,
		}
		owners = append(owners, owner)
	}

	aliases := map[string]string{}
	for k, v := range plan.CATemplateAliases.Elements() {
		aliases[k] = strings.Trim(v.String(), `"`)
	}

	application := tlspc.Application{
		ID:                   state.ID.ValueString(),
		Name:                 plan.Name.ValueString(),
		Owners:               owners,
		CertificateTemplates: aliases,
	}

	updated, err := r.client.UpdateApplication(application)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating application",
			"Could not update application, unexpected error: "+err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(updated.ID)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *applicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var plan, state applicationResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteApplication(state.ID.ValueString())
	if err != nil {
		// Just take the error out for now, at least until we try below...
		// resp.Diagnostics.AddError(
		// 	"Error Deleting Application",
		// 	"Could not delete Application ID "+state.ID.ValueString()+": "+err.Error(),
		// )

		// TODO: determine the error code here to kick in the next bit of logic.
		// Just assume whatever the error, lets update the app anyway to remove CA Templates.
		owners := []tlspc.OwnerAndType{}
		for _, v := range state.Owners {
			m := v.Elements()
			// TODO: Work out how you're supposed to get an unquoted string out
			kind := strings.Trim(m["type"].String(), `"`)
			ownerId := strings.Trim(m["owner"].String(), `"`)
			if kind != "USER" && kind != "TEAM" {
				resp.Diagnostics.AddError(
					"Error creating application",
					"Could not create application, unsupported owner type: "+kind,
				)
				return
			}
			if ownerId == "" {
				resp.Diagnostics.AddError(
					"Error creating application",
					"Could not create application, undefined owner",
				)
				return
			}
			owner := tlspc.OwnerAndType{
				ID:   ownerId,
				Type: kind,
			}
			owners = append(owners, owner)
		}

		// Set this as: {"":""} to use to overwrite the state
		aliases := map[string]string{}

		// for k, v := range state.CATemplateAliases.Elements() {
		// 	aliases[k] = strings.Trim(v.String(), `"`)
		// }

		application := tlspc.Application{
			ID:                   state.ID.ValueString(),
			Name:                 state.Name.ValueString(),
			Owners:               owners,
			CertificateTemplates: aliases,
		}

		updated, err := r.client.UpdateApplication(application)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating application",
				"Could not update application, unexpected error: "+err.Error(),
			)
			return
		}

		diags := req.State.Get(ctx, &state)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		errr := r.client.DeleteApplication(state.ID.ValueString())
		if errr != nil {
			resp.Diagnostics.AddError(
				"Error Deleting Application",
				"Could not delete Application ID "+state.ID.ValueString()+": "+errr.Error(),
			)
		}

		plan.ID = types.StringValue(updated.ID)
		// diags = resp.State.Set(ctx, plan)
		resp.Diagnostics.Append(diags...)
		return
	}
}

func (r *applicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
