# Terraform Provider for OpenRouter

Terraform Plugin Framework provider for OpenRouter management APIs.

## Supports

Resources:

- `openrouter_workspace`
- `openrouter_api_key`
- `openrouter_guardrail`

Data sources:

- `openrouter_providers`
- `openrouter_models`
- `openrouter_workspaces`

> API key safety: OpenRouter's one-time plaintext key is discarded on create and is never stored in Terraform state.

## Quick start

```shell
export OPENROUTER_MANAGEMENT_API_KEY="..."
```

```hcl
terraform {
  required_providers {
    openrouter = {
      source  = "openrouter/openrouter"
      version = "~> 0.1"
    }
  }
}

provider "openrouter" {
  # Optional; defaults to https://openrouter.ai/api/v1
  # base_url = "https://openrouter.ai/api/v1"

  user_agent = "my-platform/1.0"
}
```

## Minimal example

```hcl
resource "openrouter_workspace" "prod" {
  name = "Production"
  slug = "production"
}

resource "openrouter_api_key" "ci" {
  name         = "ci"
  workspace_id = openrouter_workspace.prod.id
  limit        = 50
  limit_reset  = "monthly"
}

resource "openrouter_guardrail" "prod" {
  name              = "production"
  workspace_id      = openrouter_workspace.prod.id
  allowed_providers = ["openai", "anthropic"]
  limit_usd         = 100
  reset_interval    = "monthly"
}

data "openrouter_models" "all" {}
```

## Import

```shell
terraform import openrouter_workspace.prod <workspace-id-or-slug>
terraform import openrouter_api_key.ci <api-key-hash>
terraform import openrouter_guardrail.prod <guardrail-id>
```

## Development

Requirements: Terraform >= 1.6, Go >= 1.24.

```shell
go test ./...
go build -o terraform-provider-openrouter
```

For local Terraform testing, add a dev override in `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "openrouter/openrouter" = "/absolute/path/to/openrouter-terraform"
  }
  direct {}
}
```

Docs live in `docs/`; examples live in `examples/`. Releases use `.goreleaser.yml` and `terraform-registry-manifest.json`.

## License

Apache License 2.0. Copyright 2026 Jonatan Jonasson.
