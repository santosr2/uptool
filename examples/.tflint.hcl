# Example TFLint configuration showing plugin version management with uptool

config {
  module = true
  force  = false
}

# AWS Plugin
plugin "aws" {
  enabled = true
  version = "0.21.0"
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}

# Azure Plugin
plugin "azurerm" {
  enabled = false
  version = "0.20.0"
  source  = "github.com/terraform-linters/tflint-ruleset-azurerm"
}

# Google Cloud Plugin
plugin "google" {
  enabled = false
  version = "0.18.0"
  source  = "github.com/terraform-linters/tflint-ruleset-google"
}

# Rule configurations
rule "aws_instance_invalid_type" {
  enabled = true
}

rule "aws_instance_previous_type" {
  enabled = true
}

rule "terraform_deprecated_index" {
  enabled = true
}

rule "terraform_unused_declarations" {
  enabled = true
}

rule "terraform_comment_syntax" {
  enabled = true
}

rule "terraform_documented_outputs" {
  enabled = true
}

rule "terraform_documented_variables" {
  enabled = true
}

rule "terraform_naming_convention" {
  enabled = true
  format  = "snake_case"
}
