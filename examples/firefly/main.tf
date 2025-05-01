terraform {
  required_providers {
    tlspc = {
      source = "jetstack/tlspc"
    }
  }
}

provider "tlspc" {
}

data "tlspc_user" "owner" {
  email = "adrian.lai@venafi.com"
}

resource "tlspc_team" "team" {
  name   = "FF Team 1"
  role   = "PLATFORM_ADMIN"
  owners = [data.tlspc_user.owner.id]
}

resource "tls_private_key" "rsa-key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "tlspc_service_account" "sa" {
  name                = "wibble-cluster"
  owner               = resource.tlspc_team.team.id
  scopes              = ["distributed-issuance"]
  credential_lifetime = 365
  public_key          = trimspace(resource.tls_private_key.rsa-key.public_key_pem)
}

resource "tlspc_firefly_policy" "ff_policy" {
  name                = "Firefly Policy 1"
  extended_key_usages = ["ANY"]
  key_usages          = ["digitalSignature", "keyEncipherment"]
  validity_period     = "P30D"
  key_algorithm = {
    allowed_values = ["RSA_2048"]
    default_value  = "RSA_2048"
  }
  sans = {
    dns_names = {
      type            = "OPTIONAL"
      min_occurrences = 0
      max_occurrences = 1000
      allowed_values  = []
      default_values  = []
    }
    ip_addresses = {
      type            = "OPTIONAL"
      min_occurrences = 0
      max_occurrences = 1000
      allowed_values  = []
      default_values  = []
    }
    rfc822_names = {
      type            = "OPTIONAL"
      min_occurrences = 0
      max_occurrences = 1000
      allowed_values  = []
      default_values  = []
    }
    uris = {
      type            = "OPTIONAL"
      min_occurrences = 0
      max_occurrences = 1000
      allowed_values  = []
      default_values  = []
    }
  }
  subject = {
    country = {
      type            = "OPTIONAL"
      min_occurrences = 0
      max_occurrences = 1000
      allowed_values  = []
      default_values  = []
    }
    common_name = {
      type            = "OPTIONAL"
      min_occurrences = 0
      max_occurrences = 1000
      allowed_values  = []
      default_values  = []
    }
    locality = {
      type            = "OPTIONAL"
      min_occurrences = 0
      max_occurrences = 1000
      allowed_values  = []
      default_values  = []
    }
    organization = {
      type            = "OPTIONAL"
      min_occurrences = 0
      max_occurrences = 1000
      allowed_values  = []
      default_values  = []
    }
    organizational_unit = {
      type            = "OPTIONAL"
      min_occurrences = 0
      max_occurrences = 1000
      allowed_values  = []
      default_values  = []
    }
    state_or_province = {
      type            = "OPTIONAL"
      min_occurrences = 0
      max_occurrences = 1000
      allowed_values  = []
      default_values  = []
    }
  }
}

data "tlspc_ca_product" "built_in_ca" {
  type           = "BUILTIN"
  ca_name        = "Built-In CA"
  product_option = "Default Product"
}

resource "tlspc_firefly_subca" "subca" {
  name                 = "Firefly Sub CA"
  ca_type              = data.tlspc_ca_product.built_in_ca.type
  ca_account_id        = data.tlspc_ca_product.built_in_ca.account_id
  ca_product_option_id = data.tlspc_ca_product.built_in_ca.id
  common_name          = "foobar"
  key_algorithm        = "RSA_2048"
  validity_period      = "P30D"
}

resource "tlspc_firefly_config" "ff_config" {
  name             = "Firefly Config"
  subca_provider   = resource.tlspc_firefly_subca.subca.id
  service_accounts = [resource.tlspc_service_account.sa.id]
  policies         = [resource.tlspc_firefly_policy.ff_policy.id]
}
