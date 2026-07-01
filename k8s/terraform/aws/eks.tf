module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 21.0"

  name               = "${local.name}-eks"
  kubernetes_version = var.kubernetes_version

  endpoint_public_access = true

  # API auth mode + let the state's caller manage the cluster; deploy roles are
  # granted access via aws_eks_access_entry (access.tf).
  enable_cluster_creator_admin_permissions = true

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  eks_managed_node_groups = {
    default = {
      instance_types = local.cfg.node_types
      desired_size   = local.cfg.node_desired
      min_size       = local.cfg.node_min
      max_size       = local.cfg.node_max
    }
  }

  tags = local.tags
}
