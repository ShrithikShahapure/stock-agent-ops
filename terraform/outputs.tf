output "cluster_name" {
  description = "Name of the EKS cluster."
  value       = module.eks.cluster_name
}

output "cluster_endpoint" {
  description = "API server endpoint URL of the EKS cluster."
  value       = module.eks.cluster_endpoint
}

output "cluster_ca_certificate" {
  description = "Base64-encoded CA certificate for the EKS cluster."
  value       = module.eks.cluster_ca_certificate
  sensitive   = true
}

output "ecr_api_url" {
  description = "ECR repository URL for the API image."
  value       = module.ecr.api_repository_url
}

output "ecr_frontend_url" {
  description = "ECR repository URL for the frontend image."
  value       = module.ecr.frontend_repository_url
}

output "ci_role_arn" {
  description = "ARN of the IAM role assumed by GitHub Actions."
  value       = module.iam.github_actions_ci_role_arn
}

output "terraform_cloud_role_arn" {
  description = "ARN of the IAM role assumed by Terraform Cloud via OIDC."
  value       = module.iam.terraform_cloud_role_arn
}

output "vpc_id" {
  description = "ID of the VPC."
  value       = module.vpc.vpc_id
}

output "public_subnet_ids" {
  description = "IDs of the public subnets."
  value       = module.vpc.public_subnet_ids
}

output "private_subnet_ids" {
  description = "IDs of the private subnets."
  value       = module.vpc.private_subnet_ids
}
