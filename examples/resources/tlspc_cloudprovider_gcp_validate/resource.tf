resource "tlspc_cloudprovider_gcp" "gcp-cloudprovider" {
  name                               = "terraform-wif"
  team                               = resource.tlspc_team.team.id
  service_account_email              = resource.google_service_account.tlspc.email
  project_number                     = data.google_project.project.number
  workload_identity_pool_id          = resource.google_iam_workload_identity_pool.tlspc.workload_identity_pool_id
  workload_identity_pool_provider_id = "venafi-provider"
}

resource "google_iam_workload_identity_pool_provider" "tlspc" {
  workload_identity_pool_id          = resource.google_iam_workload_identity_pool.tlspc.workload_identity_pool_id
  workload_identity_pool_provider_id = resource.tlspc_cloudprovider_gcp.gcp-cloudprovider.workload_identity_pool_provider_id
  display_name                       = "Venafi TLSPC"
  description                        = "Venafi WIF Pool Provider"
  attribute_mapping = {
    "google.subject" = "assertion.sub"
  }
  oidc {
    issuer_uri = resource.tlspc_cloudprovider_gcp.gcp-cloudprovider.issuer_url
  }
}

resource "tlspc_cloudprovider_gcp_validate" "gcp-cloudprovider-validation" {
  cloudprovider_id = resource.tlspc_cloudprovider_gcp.gcp-cloudprovider.id
  validate         = true
  # pool provider required to be present for successful connection validation
  depends_on = [
    resource.google_iam_workload_identity_pool_provider.tlspc
  ]
}
