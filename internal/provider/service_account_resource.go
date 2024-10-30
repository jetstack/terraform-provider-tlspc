// Copyright (c) Venafi, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"terraform-provider-tlspc/internal/tlspc"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &serviceAccountResource{}
	_ resource.ResourceWithConfigure = &serviceAccountResource{}
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
			"public_key": schema.StringAttribute{
				Required: true,
			},
			/*
				"privateKey": schema.StringAttribute{
					Required: true,
				},
			*/
			"credential_lifetime": schema.Int32Attribute{
				Required: true,
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
	ID        types.String   `tfsdk:"id"`
	Name      types.String   `tfsdk:"name"`
	Owner     types.String   `tfsdk:"owner"`
	Scopes    []types.String `tfsdk:"scopes"`
	PublicKey types.String   `tfsdk:"public_key"`
	//PrivateKey         types.String   `tfsdk:"privateKey"`
	CredentialLifetime types.Int32 `tfsdk:"credential_lifetime"`
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
		Name:               plan.Name.ValueString(),
		Owner:              plan.Owner.ValueString(),
		Scopes:             scopes,
		PublicKey:          plan.PublicKey.ValueString(),
		CredentialLifetime: plan.CredentialLifetime.ValueInt32(),
		AuthenticationType: "rsaKey",
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
}

func (r *serviceAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *serviceAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
