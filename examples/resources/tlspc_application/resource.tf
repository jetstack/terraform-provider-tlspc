resource "tlspc_application" "app" {
  name                = "TF Managed App"
  owners              = [{ type = "USER", owner = data.tlspc_user.owner.id }, { type = "TEAM", owner = resource.tlspc_team.team.id }]
  ca_template_aliases = { "${resource.tlspc_certificate_template.built_in.name}" = resource.tlspc_certificate_template.built_in.id }
}
