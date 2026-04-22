data "openrouter_models" "programming" {
  category = "programming"
}

output "model_ids" {
  value = [for model in data.openrouter_models.programming.models : model.id]
}
