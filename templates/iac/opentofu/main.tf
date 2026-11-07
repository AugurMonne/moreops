terraform {
  required_version = ">= 1.6.0"
}

locals {
  app_name = "app"
}

output "status" {
  value = "OpenTofu infrastructure ready"
}
