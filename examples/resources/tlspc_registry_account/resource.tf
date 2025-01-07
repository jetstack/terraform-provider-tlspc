resource "tlspc_registry_account" "oci" {
  name                = "k8s-oci"
  owner               = resource.tlspc_team.team.id
  scopes              = ["oci-registry-cm"]
  credential_lifetime = 365
}

output "dockerconfig" {
  value     = jsonencode({ "auths" = { "private-registry.venafi.eu" = { "auth" = base64encode("${resource.tlspc_registry_account.oci.oci_account_name}:${resource.tlspc_registry_account.oci.oci_registry_token}") } } })
  sensitive = true
}
