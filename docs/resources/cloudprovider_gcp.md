---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tlspc_cloudprovider_gcp Resource - tlspc"
subcategory: ""
description: |-
  Configure a GCP Cloud Provider integration with Workload Identity Federation
---

# tlspc_cloudprovider_gcp (Resource)

Configure a GCP Cloud Provider integration with Workload Identity Federation

## Example Usage

```terraform
provider "google" {
  project = "$PROJECT_NAME"
  region  = "europe-west1"
}

resource "google_project_iam_custom_role" "tlspc" {
  role_id     = "tlspc_wif"
  title       = "TLSPC Integration"
  description = "Permissions granted to TLSPC"
  permissions = [
    "certificatemanager.certs.create",
    "certificatemanager.certs.get",
    "certificatemanager.certs.list",
    "certificatemanager.certs.update",
    "certificatemanager.locations.list",
    "certificatemanager.operations.get",
    "resourcemanager.projects.get"
  ]
}

resource "google_service_account" "tlspc" {
  account_id   = "venafi-tlspc-wif"
  display_name = "Venafi TLSPC Workload Identity"
}

resource "google_project_iam_member" "tlspc_wif" {
  project = "$PROJECT_NAME"
  role    = resource.google_project_iam_custom_role.tlspc.id
  member  = resource.google_service_account.tlspc.member
}

resource "google_iam_workload_identity_pool" "tlspc" {
  workload_identity_pool_id = "venafi-workload-pool"
  display_name              = "Venafi TLSPC Pool"
  description               = "Venafi Workload Identity Pool"
}

resource "google_project_service" "enable_cloud_resource_manager_api" {
  service = "cloudresourcemanager.googleapis.com"
}

data "google_project" "project" {
}

resource "google_project_iam_member" "tlspc_wi_user" {
  project = "$PROJECT_NAME"
  role    = "roles/iam.workloadIdentityUser"
  member  = "principal://iam.googleapis.com/projects/${data.google_project.project.number}/locations/global/workloadIdentityPools/${resource.google_iam_workload_identity_pool.tlspc.workload_identity_pool_id}/subject/venafi_control_plane"
}

data "tlspc_user" "owner" {
  email = "admin@admin.com"
}

resource "tlspc_team" "team" {
  name   = "TF WIF"
  role   = "PLATFORM_ADMIN"
  owners = [data.tlspc_user.owner.id]
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

resource "tlspc_cloudprovider_gcp" "gcp-cloudprovider" {
  name                               = "terraform-wif"
  team                               = resource.tlspc_team.team.id
  service_account_email              = resource.google_service_account.tlspc.email
  project_number                     = data.google_project.project.number
  workload_identity_pool_id          = resource.google_iam_workload_identity_pool.tlspc.workload_identity_pool_id
  workload_identity_pool_provider_id = "venafi-provider"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of this integration
- `project_number` (Number) GCP Project Number
- `service_account_email` (String) GCP Service Account Email
- `team` (String) The ID of the owning Team
- `workload_identity_pool_id` (String) GCP Workload Identity Pool ID
- `workload_identity_pool_provider_id` (String) GCP Workload Identity Pool Provider ID

### Read-Only

- `id` (String) The ID of this resource
- `issuer_url` (String) The issuer URL that should be provided to set up the GCP Workload Identity Pool
