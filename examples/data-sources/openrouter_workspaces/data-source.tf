data "openrouter_workspaces" "all" {}

output "workspace_slugs" {
  value = [for workspace in data.openrouter_workspaces.all.workspaces : workspace.slug]
}
