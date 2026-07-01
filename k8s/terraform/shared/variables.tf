variable "region" {
  type    = string
  default = "us-east-1"
}

variable "project" {
  type    = string
  default = "auto-repair-shop"
}

variable "github_repo" {
  description = "owner/repo allowed to assume the deploy roles via OIDC"
  type        = string
  default     = "POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop"
}

variable "environments" {
  description = "GitHub Actions environments that each get a deploy role"
  type        = list(string)
  default     = ["STG", "PRD"]
}
