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
	_ resource.Resource                = &serviceAccountResource{}
	_ resource.ResourceWithConfigure   = &serviceAccountResource{}
	_ resource.ResourceWithImportState = &serviceAccountResource{}
)

type serviceAccountResource struct {
	client *tlspc.Client
}

func NewServiceAccountResource() resource.Resource {
	return &serviceAccountResource{}
}

func (r *serviceAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (r *serviceAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			// Agent service account
			"public_key": schema.StringAttribute{
				Optional: true,
			},
			"credential_lifetime": schema.Int32Attribute{
				Optional: true,
			},
			// Issuer service account (jwks)
			"jwks_uri": schema.StringAttribute{
				Optional: true,
			},
			"issuer_url": schema.StringAttribute{
				Optional: true,
			},
			"audience": schema.StringAttribute{
				Optional: true,
			},
			"subject": schema.StringAttribute{
				Optional: true,
			},
			"applications": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *serviceAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type serviceAccountResourceModel struct {
	ID                 types.String   `tfsdk:"id"`
	Name               types.String   `tfsdk:"name"`
	Owner              types.String   `tfsdk:"owner"`
	Scopes             []types.String `tfsdk:"scopes"`
	PublicKey          types.String   `tfsdk:"public_key"`
	CredentialLifetime types.Int32    `tfsdk:"credential_lifetime"`
	JwksURI            types.String   `tfsdk:"jwks_uri"`
	IssuerURL          types.String   `tfsdk:"issuer_url"`
	Audience           types.String   `tfsdk:"audience"`
	Subject            types.String   `tfsdk:"subject"`
	Applications       []types.String `tfsdk:"applications"`
}

func (r *serviceAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceAccountResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	scopes := []string{}
	for _, v := range plan.Scopes {
		scopes = append(scopes, v.ValueString())
	}

	serviceAccount := tlspc.ServiceAccount{
		Name:   plan.Name.ValueString(),
		Owner:  plan.Owner.ValueString(),
		Scopes: scopes,
	}

	configured := false
	// Agent type
	if plan.PublicKey.ValueString() != "" || plan.CredentialLifetime.ValueInt32() > 0 {
		serviceAccount.PublicKey = plan.PublicKey.ValueString()
		serviceAccount.CredentialLifetime = plan.CredentialLifetime.ValueInt32()
		serviceAccount.AuthenticationType = "rsaKey"
		configured = true
	}

	// Issuer type
	if plan.JwksURI.ValueString() != "" || plan.IssuerURL.ValueString() != "" || plan.Audience.ValueString() != "" || plan.Subject.ValueString() != "" || len(plan.Applications) > 0 {
		if serviceAccount.AuthenticationType == "rsaKey" {
			resp.Diagnostics.AddError(
				"Error creating serviceAccount",
				"Could not create serviceAccount, invalid configuration (both public_key and jwks fields present)",
			)
			return
		}
		serviceAccount.JwksURI = plan.JwksURI.ValueString()
		serviceAccount.IssuerURL = plan.IssuerURL.ValueString()
		serviceAccount.Audience = plan.Audience.ValueString()
		serviceAccount.Subject = plan.Subject.ValueString()
		serviceAccount.AuthenticationType = "rsaKeyFederated"

		apps := []string{}
		for _, v := range plan.Applications {
			apps = append(apps, v.ValueString())
		}
		serviceAccount.Applications = apps
		configured = true
	}
	if !configured {
		resp.Diagnostics.AddError(
			"Error creating serviceAccount",
			"Could not create serviceAccount, invalid configuration (neither public_key or jwks fields present)",
		)
		return
	}

	created, err := r.client.CreateServiceAccount(serviceAccount)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating serviceAccount",
			"Could not create serviceAccount, unexpected error: "+err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(created.ID)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *serviceAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceAccountResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sa, err := r.client.GetServiceAccount(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Service Account",
			"Could not read service account ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(sa.ID)
	state.Name = types.StringValue(sa.Name)
	state.Owner = types.StringValue(sa.Owner)
	state.PublicKey = types.StringValue(sa.PublicKey)
	state.CredentialLifetime = types.Int32Value(sa.CredentialLifetime)

	scopes := []types.String{}
	for _, v := range sa.Scopes {
		scopes = append(scopes, types.StringValue(v))
	}
	state.Scopes = scopes

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *serviceAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state serviceAccountResourceModel

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

	serviceAccount := tlspc.ServiceAccount{
		ID:     state.ID.ValueString(),
		Name:   plan.Name.ValueString(),
		Owner:  plan.Owner.ValueString(),
		Scopes: scopes,
	}

	configured := false
	// Agent type
	if plan.PublicKey.ValueString() != "" || plan.CredentialLifetime.ValueInt32() > 0 {
		serviceAccount.PublicKey = plan.PublicKey.ValueString()
		serviceAccount.CredentialLifetime = plan.CredentialLifetime.ValueInt32()
		serviceAccount.AuthenticationType = "rsaKey"
		configured = true
	}

	// Issuer type
	if plan.JwksURI.ValueString() != "" || plan.IssuerURL.ValueString() != "" || plan.Audience.ValueString() != "" || plan.Subject.ValueString() != "" || len(plan.Applications) > 0 {
		if serviceAccount.AuthenticationType == "rsaKey" {
			resp.Diagnostics.AddError(
				"Error creating serviceAccount",
				"Could not create serviceAccount, invalid configuration (both public_key and jwks fields present)",
			)
			return
		}
		serviceAccount.JwksURI = plan.JwksURI.ValueString()
		if state.IssuerURL.ValueString() != plan.IssuerURL.ValueString() {
			serviceAccount.IssuerURL = plan.IssuerURL.ValueString()
		}
		serviceAccount.Audience = plan.Audience.ValueString()
		if state.Subject.ValueString() != plan.Subject.ValueString() {
			serviceAccount.Subject = plan.Subject.ValueString()
		}
		serviceAccount.AuthenticationType = "rsaKeyFederated"

		apps := []string{}
		for _, v := range plan.Applications {
			apps = append(apps, v.ValueString())
		}
		serviceAccount.Applications = apps
		configured = true
	}
	if !configured {
		resp.Diagnostics.AddError(
			"Error creating serviceAccount",
			"Could not create serviceAccount, invalid configuration (neither public_key or jwks fields present)",
		)
		return
	}

	err := r.client.UpdateServiceAccount(serviceAccount)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating serviceAccount",
			"Could not update serviceAccount, unexpected error: "+err.Error(),
		)
		return
	}
	plan.ID = state.ID
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *serviceAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceAccountResourceModel

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

func (r *serviceAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
