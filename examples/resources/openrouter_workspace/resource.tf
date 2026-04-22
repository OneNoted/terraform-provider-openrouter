resource "openrouter_workspace" "example" {
  name                  = "Example"
  slug                  = "example"
  description           = "Example workspace managed by Terraform"
  default_text_model    = "openai/gpt-4o"
  default_image_model   = "openai/dall-e-3"
  default_provider_sort = "price"
}
