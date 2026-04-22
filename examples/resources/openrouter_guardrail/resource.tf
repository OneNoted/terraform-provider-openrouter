resource "openrouter_guardrail" "example" {
  name              = "Example Guardrail"
  workspace_id      = openrouter_workspace.example.id
  description       = "Example provider and spending controls"
  allowed_providers = ["openai", "anthropic"]
  enforce_zdr       = false
  limit_usd         = 100
  reset_interval    = "monthly"
}
