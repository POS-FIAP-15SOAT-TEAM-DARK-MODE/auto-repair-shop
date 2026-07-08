module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 21.0"

  name               = "${local.name}-eks"
  kubernetes_version = var.kubernetes_version

  endpoint_public_access = true

  # API auth mode + let the state's caller manage the cluster; deploy roles are
  # granted access via aws_eks_access_entry (access.tf).
  enable_cluster_creator_admin_permissions = true

  # On restricted accounts (manage_iam = false) we cannot create IAM roles or the
  # IRSA OIDC provider, so the cluster reuses an existing role (e.g. LabRole) and
  # IRSA is turned off (no IRSA-based add-ons in that mode).
  create_iam_role = var.manage_iam
  iam_role_arn    = var.manage_iam ? null : var.execution_role_arn
  enable_irsa     = var.manage_iam

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  eks_managed_node_groups = {
    default = {
      instance_types = local.cfg.node_types
      desired_size   = local.cfg.node_desired
      min_size       = local.cfg.node_min
      max_size       = local.cfg.node_max

      # Reuse the existing role on restricted accounts.
      create_iam_role = var.manage_iam
      iam_role_arn    = var.manage_iam ? null : var.execution_role_arn
    }
  }

  tags = local.tags
}
