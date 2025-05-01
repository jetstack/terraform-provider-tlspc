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
	_ resource.Resource                = &fireflyPolicyResource{}
	_ resource.ResourceWithConfigure   = &fireflyPolicyResource{}
	_ resource.ResourceWithImportState = &fireflyPolicyResource{}
)

type fireflyPolicyResource struct {
	client *tlspc.Client
}

func NewFireflyPolicyResource() resource.Resource {
	return &fireflyPolicyResource{}
}

func (r *fireflyPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firefly_policy"
}

func (r *fireflyPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	policyAttr := schema.SingleNestedAttribute{
		Required: true,
		Attributes: map[string]schema.Attribute{
			"allowed_values": schema.SetAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: `A list of allowed values, may be literal strings or regular expressions. Regular expressions must be prefixed with '^'`,
			},
			"default_values": schema.SetAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: `A list of default values`,
			},
			"max_occurrences": schema.Int32Attribute{
				Required: true,
			},
			"min_occurrences": schema.Int32Attribute{
				Required: true,
			},
			"type": schema.StringAttribute{
				Required: true,
				MarkdownDescription: `The type of this constraint, valid options include:
	* IGNORED
	* FORBIDDEN
	* OPTIONAL
	* REQUIRED
`,
			},
		},
	}

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
				MarkdownDescription: "The name of the Firefly Policy",
			},
			"extended_key_usages": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				MarkdownDescription: `List of Extended Key usages, valid options include:
	* ANY
	* SERVER_AUTH
	* CLIENT~_AUTH
	* CODE_SIGNING
	* EMAIL_PROTECTION
	* IPSEC_ENDSYSTEM
	* IPSEC_TUNNEL
	* IPSEC_USER
	* TIME_STAMPING
	* OCSP_SIGNING
	* DVCS
	* SBGP_CERT_AA_SERVER_AUTH
	* SCVP_RESPONDER
	* EAP_OVER_PPP
	* EAP_OVER_LAN
	* SCVP_SERVER
	* SCVP_CLIENT
	* IPSEC_IKE
	* CAPWAP_AC
	* CAPWAP_WTP
	* IPSEC_IKE_INTERMEDIATE
	* SMARTCARD_LOGON
`,
			},
			"key_usages": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				MarkdownDescription: `List of Key usages, valid options include:
	* digitalSignature
	* nonRepudiation
	* keyEncipherment
	* dataEncipherment
	* keyAgreement
	* keyCertSign
	* cRLSign
	* encipherOnly
	* decipherOnly
`,
			},
			"validity_period": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Validity Period in ISO8601 Period Format. e.g. P30D",
			},
			"key_algorithm": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"allowed_values": schema.SetAttribute{
						Required:    true,
						ElementType: types.StringType,
						MarkdownDescription: `A list of allowed Key Algorithm. Valid options include:
	* RSA_2048
	* RSA_3072
	* RSA_4096
	* EC_P256
	* EC_P384
	* EC_P521
	* EC_ED25519
`,
					},
					"default_value": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: `Default key algorithm`,
					},
				},
			},
			"sans": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: `Policy for Subject Alternative Names`,
				Attributes: map[string]schema.Attribute{
					"dns_names":    policyAttr,
					"ip_addresses": policyAttr,
					"rfc822_names": policyAttr,
					"uris":         policyAttr,
				},
			},
			"subject": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: `Policy for Subject`,
				Attributes: map[string]schema.Attribute{
					"common_name":         policyAttr,
					"country":             policyAttr,
					"locality":            policyAttr,
					"organization":        policyAttr,
					"organizational_unit": policyAttr,
					"state_or_province":   policyAttr,
				},
			},
		},
	}
}

func (r *fireflyPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type fireflyPolicyResourceModel struct {
	ID                types.String      `tfsdk:"id"`
	Name              types.String      `tfsdk:"name"`
	ExtendedKeyUsages []types.String    `tfsdk:"extended_key_usages"`
	KeyUsages         []types.String    `tfsdk:"key_usages"`
	ValidityPeriod    types.String      `tfsdk:"validity_period"`
	KeyAlgorithm      keyAlgorithmModel `tfsdk:"key_algorithm"`
	SANs              sansModel         `tfsdk:"sans"`
	Subject           subjectModel      `tfsdk:"subject"`
}

type keyAlgorithmModel struct {
	AllowedValues []types.String `tfsdk:"allowed_values"`
	DefaultValue  types.String   `tfsdk:"default_value"`
}

type policyModel struct {
	AllowedValues  []types.String `tfsdk:"allowed_values"`
	DefaultValues  []types.String `tfsdk:"default_values"`
	MaxOccurrences types.Int32    `tfsdk:"max_occurrences"`
	MinOccurrences types.Int32    `tfsdk:"min_occurrences"`
	Type           types.String   `tfsdk:"type"`
}

type sansModel struct {
	DNSNames    policyModel `tfsdk:"dns_names"`
	IPAddresses policyModel `tfsdk:"ip_addresses"`
	RFC822Names policyModel `tfsdk:"rfc822_names"`
	URIs        policyModel `tfsdk:"uris"`
}

type subjectModel struct {
	CommonName         policyModel `tfsdk:"common_name"`
	Country            policyModel `tfsdk:"country"`
	Locality           policyModel `tfsdk:"locality"`
	Organization       policyModel `tfsdk:"organization"`
	OrganizationalUnit policyModel `tfsdk:"organizational_unit"`
	StateOrProvince    policyModel `tfsdk:"state_or_province"`
}

func (r *fireflyPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan fireflyPolicyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ff := coercePolicy(plan)
	created, err := r.client.CreateFireflyPolicy(ff)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Firefly Policy",
			"Could not create Firefly Policy, unexpected error: "+err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(created.ID)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func coercePolicy(plan fireflyPolicyResourceModel) tlspc.FireflyPolicy {
	extKeys := []string{}
	for _, v := range plan.ExtendedKeyUsages {
		extKeys = append(extKeys, v.ValueString())
	}

	keyAlgAllowed := []string{}
	for _, v := range plan.KeyAlgorithm.AllowedValues {
		keyAlgAllowed = append(keyAlgAllowed, v.ValueString())
	}
	keyAlg := tlspc.KeyAlgorithm{
		DefaultValue:  plan.KeyAlgorithm.DefaultValue.ValueString(),
		AllowedValues: keyAlgAllowed,
	}

	keyUses := []string{}
	for _, v := range plan.KeyUsages {
		keyUses = append(keyUses, v.ValueString())
	}

	return tlspc.FireflyPolicy{
		Name:              plan.Name.ValueString(),
		ExtendedKeyUsages: extKeys,
		KeyAlgorithm:      keyAlg,
		KeyUsages:         keyUses,
		SANs: tlspc.SANs{
			DNSNames:    coercePolicyDetails(plan.SANs.DNSNames),
			IPAddresses: coercePolicyDetails(plan.SANs.IPAddresses),
			RFC822Names: coercePolicyDetails(plan.SANs.RFC822Names),
			URIs:        coercePolicyDetails(plan.SANs.URIs),
		},
		Subject: tlspc.FireflyPolicySubject{
			CommonName:         coercePolicyDetails(plan.Subject.CommonName),
			Country:            coercePolicyDetails(plan.Subject.Country),
			Locality:           coercePolicyDetails(plan.Subject.Locality),
			Organization:       coercePolicyDetails(plan.Subject.Organization),
			OrganizationalUnit: coercePolicyDetails(plan.Subject.OrganizationalUnit),
			StateOrProvince:    coercePolicyDetails(plan.Subject.StateOrProvince),
		},
		ValidityPeriod: plan.ValidityPeriod.ValueString(),
	}
}

func coercePolicyDetails(p policyModel) tlspc.PolicyDetails {
	av := []string{}
	for _, v := range p.AllowedValues {
		av = append(av, v.ValueString())
	}

	dv := []string{}
	for _, v := range p.DefaultValues {
		dv = append(dv, v.ValueString())
	}

	return tlspc.PolicyDetails{
		AllowedValues:  av,
		DefaultValues:  dv,
		MaxOccurrences: p.MaxOccurrences.ValueInt32(),
		MinOccurrences: p.MinOccurrences.ValueInt32(),
		Type:           p.Type.ValueString(),
	}

}

func coercePolicyModel(p tlspc.PolicyDetails) policyModel {
	av := []types.String{}
	for _, v := range p.AllowedValues {
		av = append(av, types.StringValue(v))
	}

	dv := []types.String{}
	for _, v := range p.DefaultValues {
		dv = append(dv, types.StringValue(v))
	}

	return policyModel{
		AllowedValues:  av,
		DefaultValues:  dv,
		MaxOccurrences: types.Int32Value(p.MaxOccurrences),
		MinOccurrences: types.Int32Value(p.MinOccurrences),
		Type:           types.StringValue(p.Type),
	}
}

func (r *fireflyPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state fireflyPolicyResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ff, err := r.client.GetFireflyPolicy(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading FireflyConfig",
			"Could not read FireflyConfig ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(ff.ID)
	state.Name = types.StringValue(ff.Name)
	state.ValidityPeriod = types.StringValue(ff.ValidityPeriod)

	extKeys := []types.String{}
	for _, v := range ff.ExtendedKeyUsages {
		extKeys = append(extKeys, types.StringValue(v))
	}
	state.ExtendedKeyUsages = extKeys

	keyUses := []types.String{}
	for _, v := range ff.KeyUsages {
		keyUses = append(keyUses, types.StringValue(v))
	}
	state.KeyUsages = keyUses

	allowed := []types.String{}
	for _, v := range ff.KeyAlgorithm.AllowedValues {
		allowed = append(allowed, types.StringValue(v))
	}
	state.KeyAlgorithm = keyAlgorithmModel{
		AllowedValues: allowed,
		DefaultValue:  types.StringValue(ff.KeyAlgorithm.DefaultValue),
	}

	state.SANs = sansModel{
		DNSNames:    coercePolicyModel(ff.SANs.DNSNames),
		IPAddresses: coercePolicyModel(ff.SANs.IPAddresses),
		RFC822Names: coercePolicyModel(ff.SANs.RFC822Names),
		URIs:        coercePolicyModel(ff.SANs.URIs),
	}

	state.Subject = subjectModel{
		CommonName:         coercePolicyModel(ff.Subject.CommonName),
		Country:            coercePolicyModel(ff.Subject.Country),
		Locality:           coercePolicyModel(ff.Subject.Locality),
		Organization:       coercePolicyModel(ff.Subject.Organization),
		OrganizationalUnit: coercePolicyModel(ff.Subject.OrganizationalUnit),
		StateOrProvince:    coercePolicyModel(ff.Subject.StateOrProvince),
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *fireflyPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state fireflyPolicyResourceModel

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

	ff := coercePolicy(plan)
	ff.ID = state.ID.ValueString()

	updated, err := r.client.UpdateFireflyPolicy(ff)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Firefly Policy",
			"Could not update Firefly Policy, unexpected error: "+err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(updated.ID)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *fireflyPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state fireflyPolicyResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteFireflyPolicy(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Firefly Policy ",
			"Could not delete Firefly Policy ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *fireflyPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
