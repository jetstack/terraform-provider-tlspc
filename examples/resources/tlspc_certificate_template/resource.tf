resource "tlspc_certificate_template" "built_in" {
  name          = "Built-In CA Cert Template"
  ca_type       = data.tlspc_ca_product.built_in.type
  ca_product_id = data.tlspc_ca_product.built_in.id
  key_reuse     = false
}
