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
	_ resource.Resource                = &certificateTemplateResource{}
	_ resource.ResourceWithConfigure   = &certificateTemplateResource{}
	_ resource.ResourceWithImportState = &certificateTemplateResource{}
)

type certificateTemplateResource struct {
	client *tlspc.Client
}

func NewCertificateTemplateResource() resource.Resource {
	return &certificateTemplateResource{}
}

func (r *certificateTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate_template"
}

func (r *certificateTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manage Certificate Issuing Template

-> Currently only a limited subset of attributes are supported. All Common Name/SAN/CSR validation fields are set to ` + "`.*` (allow all)." + ` Permitted Key Algorithms are set to RSA 2048/3072/4096.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the Certificate Issuing Template",
			},
			"ca_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Type of Certificate Authority (see Certificate Authority Product Option data source)",
			},
			"ca_product_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of a Certificate Authority Product Option",
			},
			"key_reuse": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Allow Private Key Reuse",
			},
			/*
				"key_types": schema.SetAttribute{
					Required:    true,
					ElementType: types.MapType,
				},
			*/
		},
	}
}

func (r *certificateTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type certificateTemplateResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	CAType      types.String `tfsdk:"ca_type"`
	CAProductID types.String `tfsdk:"ca_product_id"`
	KeyReuse    types.Bool   `tfsdk:"key_reuse"`
	//KeyTypes    []types.Map  `tfsdk:"key_types"`
}

func (r *certificateTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan certificateTemplateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	pt, err := r.client.GetCAProductOptionByID(plan.CAType.ValueString(), plan.CAProductID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating certificate template",
			"CA Product ID not found: "+err.Error(),
		)
		return
	}

	ct := tlspc.CertificateTemplate{
		Name:                                plan.Name.ValueString(),
		CertificateAuthorityType:            plan.CAType.ValueString(),
		CertificateAuthorityProductOptionID: plan.CAProductID.ValueString(),
		Product:                             pt.Details.Template,
		KeyReuse:                            plan.KeyReuse.ValueBool(),
		KeyTypes: []tlspc.KeyType{
			{
				Type:       "RSA",
				KeyLengths: []int32{2048, 3072, 4096},
			},
		},
		SANRegexes:       []string{".*"},
		SubjectCNRegexes: []string{".*"},
		SubjectCValues:   []string{".*"},
		SubjectLRegexes:  []string{".*"},
		SubjectORegexes:  []string{".*"},
		SubjectOURegexes: []string{".*"},
		SubjectSTRegexes: []string{".*"},
	}

	created, err := r.client.CreateCertificateTemplate(ct)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating certificate template",
			"Could not create certificate template, unexpected error: "+err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(created.ID)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *certificateTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state certificateTemplateResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ct, err := r.client.GetCertificateTemplate(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Certificate Template",
			"Could not read Certificate Template ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(ct.ID)
	state.CAType = types.StringValue(ct.CertificateAuthorityType)
	state.CAProductID = types.StringValue(ct.CertificateAuthorityProductOptionID)
	state.KeyReuse = types.BoolValue(ct.KeyReuse)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *certificateTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state certificateTemplateResourceModel

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

	pt, err := r.client.GetCAProductOptionByID(plan.CAType.ValueString(), plan.CAProductID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating certificate template",
			"CA Product ID not found: "+err.Error(),
		)
		return
	}

	ct := tlspc.CertificateTemplate{
		ID:                                  state.ID.ValueString(),
		Name:                                plan.Name.ValueString(),
		CertificateAuthorityType:            plan.CAType.ValueString(),
		CertificateAuthorityProductOptionID: plan.CAProductID.ValueString(),
		Product:                             pt.Details.Template,
		KeyReuse:                            plan.KeyReuse.ValueBool(),
		KeyTypes: []tlspc.KeyType{
			{
				Type:       "RSA",
				KeyLengths: []int32{2048, 3072, 4096},
			},
		},
		SANRegexes:       []string{".*"},
		SubjectCNRegexes: []string{".*"},
		SubjectCValues:   []string{".*"},
		SubjectLRegexes:  []string{".*"},
		SubjectORegexes:  []string{".*"},
		SubjectOURegexes: []string{".*"},
		SubjectSTRegexes: []string{".*"},
	}

	updated, err := r.client.UpdateCertificateTemplate(ct)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating certificate template",
			"Could not update certificate template, unexpected error: "+err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(updated.ID)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *certificateTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state certificateTemplateResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteCertificateTemplate(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Certificate Template",
			"Could not delete Certificate Template ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *certificateTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
