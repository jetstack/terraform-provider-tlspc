resource "tlspc_plugin" "digicert" {
  type     = "CA"
  manifest = file("${path.root}/plugins/digicert.json")
}
