output "github_actions_ci_role_arn" {
  description = "ARN of the IAM role assumed by GitHub Actions via OIDC."
  value       = aws_iam_role.github_actions_ci.arn
}

output "terraform_cloud_role_arn" {
  description = "ARN of the IAM role assumed by Terraform Cloud via OIDC."
  value       = aws_iam_role.terraform_cloud.arn
}

output "eks_cluster_role_arn" {
  description = "ARN of the IAM role used by the EKS control plane."
  value       = aws_iam_role.eks_cluster.arn
}

output "eks_node_role_arn" {
  description = "ARN of the IAM role used by EKS worker nodes."
  value       = aws_iam_role.eks_node.arn
}
