---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tlspc_service_account Resource - tlspc"
subcategory: ""
description: |-
  
---

# tlspc_service_account (Resource)



## Example Usage

```terraform
resource "tls_private_key" "rsa-key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "tlspc_service_account" "agent-credentials" {
  name                = "k8s-cluster"
  owner               = resource.tlspc_team.team.id
  scopes              = ["kubernetes-discovery"]
  credential_lifetime = 365
  public_key          = trimspace(resource.tls_private_key.rsa-key.public_key_pem)
}

resource "kubernetes_secret" "credentials" {
  metadata {
    name      = "agent-credentials"
    namespace = "venafi"
  }
  data = {
    "privatekey.pem" = tls_private_key.rsa-key.private_key_pem
  }
  type = "kubernetes.io/opaque"
}

resource "tlspc_service_account" "wif-issuer" {
  name         = "test-issuer1"
  owner        = resource.tlspc_team.team.id
  scopes       = ["certificate-issuance"]
  applications = [resource.tlspc_application.app.id]
  jwks_uri     = "https://kubernetes/.well-known/jwks.json"
  issuer_url   = "https://kubernetes.default.svc.cluster.local"
  subject      = "system:serviceaccount:venafi:application-team-1"
  audience     = "api.venafi.eu"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the service account
- `owner` (String) ID of the team that owns this service account
- `scopes` (Set of String) A list of scopes that this service account is authorised for. Available options include:
    * certificate-issuance
    * kubernetes-discovery

### Optional

- `applications` (Set of String) List of Applications which this service account is authorised for
- `audience` (String) Audience for a WIF type service account
- `credential_lifetime` (Number) Credential Lifetime in days (required for public_key type service accounts)
- `issuer_url` (String) Issuer URL for a WIF type service account
- `jwks_uri` (String) The JWKS URI for a Workload Identity Federation (WIF) type service account
- `public_key` (String) Public Key
- `subject` (String) Subject for a WIF type service account

### Read-Only

- `id` (String) The ID of this resource
