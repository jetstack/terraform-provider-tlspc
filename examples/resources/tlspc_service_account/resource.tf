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
