terraform {
  required_providers {
    tlspc = {
      source = "venafi.com/dev/tlspc"
    }
  }
}

provider "tlspc" {
  apikey   = ""
  endpoint = "https://api.venafi.eu"
}

resource "tls_private_key" "rsa-key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

output "rsa_key_public" {
  value = resource.tls_private_key.rsa-key.public_key_pem
}

data "tlspc_user" "owner" {
  email = "username@venafi.com"
}

output "owner_id" {
  value = data.tlspc_user.owner.id
}

resource "tlspc_team" "team" {
  name   = "Team TF"
  role   = "PLATFORM_ADMIN"
  owners = [data.tlspc_user.owner.id]
}

resource "tlspc_service_account" "sa" {
  name                = "example-cluster"
  owner               = resource.tlspc_team.team.id
  scopes              = ["kubernetes-discovery"]
  credential_lifetime = 365
  public_key          = resource.tls_private_key.rsa-key.public_key_pem
}

output "service_account" {
  value = resource.tlspc_service_account.sa
}
