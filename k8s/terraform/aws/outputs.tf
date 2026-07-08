output "environment" {
  value = local.environment
}

output "region" {
  value = var.region
}

output "cluster_name" {
  value = local.cluster_name
}

output "cluster_endpoint" {
  value = local.cluster_endpoint
}

output "vpc_id" {
  value = module.vpc.vpc_id
}

output "oidc_provider_arn" {
  description = "Cluster OIDC provider ARN, for IRSA in the addons state. Empty when manage_iam = false."
  value       = local.oidc_provider_arn
}

output "db_host" {
  description = "RDS endpoint; set as POSTGRES_HOST in the env overlay"
  value       = module.rds.db_instance_address
}

output "db_port" {
  value = module.rds.db_instance_port
}

output "app_secret_arn" {
  description = "Secrets Manager secret with POSTGRES_PASSWORD/JWT_SECRET (for External Secrets)"
  value       = aws_secretsmanager_secret.app.arn
}

output "configure_kubectl" {
  description = "Point kubectl at this cluster"
  value       = "aws eks update-kubeconfig --region ${var.region} --name ${local.cluster_name}"
}
