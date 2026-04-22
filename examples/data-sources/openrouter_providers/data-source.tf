data "openrouter_providers" "all" {}

output "provider_slugs" {
  value = [for provider in data.openrouter_providers.all.providers : provider.slug]
}
