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

# ── Managed services ─────────────────────────────────────────────────────────

output "elasticache_endpoint" {
  description = "ElastiCache Redis primary endpoint."
  value       = module.elasticache.primary_endpoint
}

output "sqs_queue_url" {
  description = "SQS training job queue URL."
  value       = module.sqs.queue_url
}

output "sqs_dlq_url" {
  description = "SQS dead-letter queue URL."
  value       = module.sqs.dlq_url
}

output "secrets_arn" {
  description = "ARN of the Secrets Manager secret."
  value       = module.secrets.secret_arn
}

output "secrets_name" {
  description = "Name of the Secrets Manager secret."
  value       = module.secrets.secret_name
}

# ── SageMaker ─────────────────────────────────────────────────────────────────

output "sagemaker_role_arn" {
  description = "ARN of the SageMaker execution role."
  value       = module.sagemaker.execution_role_arn
}

output "sagemaker_bucket" {
  description = "S3 bucket for SageMaker training artifacts."
  value       = module.sagemaker.bucket_name
}
