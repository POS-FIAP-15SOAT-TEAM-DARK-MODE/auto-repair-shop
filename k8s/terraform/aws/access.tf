# Pull the GitHub Actions deploy role ARNs from the shared stack.
data "terraform_remote_state" "shared" {
  backend = "s3"
  config = {
    bucket = "auto-repair-shop-tfstate-4c2c18c0"
    key    = "shared/terraform.tfstate"
    region = var.region
  }
}

locals {
  deploy_role_arn = data.terraform_remote_state.shared.outputs.deploy_role_arns[upper(local.environment)]
}

# Grant this environment's deploy role edit access to this cluster, so the
# pipeline can `kubectl apply` after `aws eks update-kubeconfig`.
resource "aws_eks_access_entry" "deploy" {
  cluster_name  = module.eks.cluster_name
  principal_arn = local.deploy_role_arn
  type          = "STANDARD"
}

resource "aws_eks_access_policy_association" "deploy" {
  cluster_name  = module.eks.cluster_name
  principal_arn = local.deploy_role_arn
  policy_arn    = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSEditPolicy"

  access_scope {
    type = "cluster"
  }
}
