resource "openrouter_api_key" "example" {
  name                  = "Example Terraform Key"
  workspace_id          = openrouter_workspace.example.id
  limit                 = 25
  limit_reset           = "monthly"
  include_byok_in_limit = true
  disabled              = false
}
