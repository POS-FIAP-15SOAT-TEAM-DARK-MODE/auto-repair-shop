variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "project" {
  description = "Project name, used as a resource prefix"
  type        = string
  default     = "auto-repair-shop"
}

# --- Network ---
variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

# --- EKS ---
variable "kubernetes_version" {
  description = "EKS control plane version"
  type        = string
  default     = "1.31"
}

# --- RDS (Postgres) ---
# Per-environment sizing (instance class, multi-AZ) is derived from the
# workspace in locals.tf; these are the values common to all environments.
variable "db_name" {
  type    = string
  default = "autorepairshop"
}

variable "db_username" {
  type    = string
  default = "postgres"
}

variable "db_allocated_storage" {
  type    = number
  default = 20
}

variable "db_engine_version" {
  type    = string
  default = "15"
}
