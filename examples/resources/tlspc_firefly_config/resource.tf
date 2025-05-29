resource "tlspc_firefly_config" "ff_config" {
  name             = "Firefly Config"
  subca_provider   = resource.tlspc_firefly_subca.subca.id
  service_accounts = [resource.tlspc_service_account.sa.id]
  policies         = [resource.tlspc_firefly_policy.ff_policy.id]
}
