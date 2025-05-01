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
