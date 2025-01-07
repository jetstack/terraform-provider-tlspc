resource "tls_private_key" "rsa-key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "tlspc_service_account" "sa" {
  name                = "k8s-cluster"
  owner               = resource.tlspc_team.team.id
  scopes              = ["kubernetes-discovery"]
  credential_lifetime = 365
  public_key          = trimspace(resource.tls_private_key.rsa-key.public_key_pem)
}
