terraform {
  required_providers {
    openrouter = {
      source = "openrouter/openrouter"
    }
  }
}

provider "openrouter" {
  # Prefer OPENROUTER_MANAGEMENT_API_KEY instead of inline secrets.
  user_agent = "terraform-provider-openrouter-example/1.0"
}
