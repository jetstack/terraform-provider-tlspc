resource "tlspc_team" "app_team_1" {
  name   = "App Team 1"
  role   = "PLATFORM_ADMIN"
  owners = [data.tlspc_user.owner.id]
}
