terraform {
  required_version = ">= 1.10"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }

  # Remote state; stg/prd live in separate Terraform workspaces, stored under
  # the env:/<workspace>/ prefix of the same bucket/key.
  backend "s3" {
    bucket       = "auto-repair-shop-tfstate"
    key          = "aws/terraform.tfstate"
    region       = "us-east-1"
    encrypt      = true
    use_lockfile = true
  }
}

provider "aws" {
  region = var.region
}
