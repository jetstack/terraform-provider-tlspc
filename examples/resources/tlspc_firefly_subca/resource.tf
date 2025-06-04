resource "tlspc_firefly_subca" "subca" {
  name                 = "Firefly Sub CA"
  ca_type              = data.tlspc_ca_product.built_in_ca.type
  ca_account_id        = data.tlspc_ca_product.built_in_ca.account_id
  ca_product_option_id = data.tlspc_ca_product.built_in_ca.id
  common_name          = "firefly-subca.com"
  key_algorithm        = "RSA_2048"
  validity_period      = "P30D"
}
