output "function_name" {
  description = "Name of the Lambda function."

  value = module.app.function_name
}

output "base_url" {
  description = "Base URL for API Gateway stage."

  value = module.app.base_url
}