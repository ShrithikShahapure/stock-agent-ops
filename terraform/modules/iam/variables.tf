variable "cluster_name" {
  description = "EKS cluster name — used to scope IAM role names and EKS ARNs."
  type        = string
}

variable "aws_region" {
  description = "AWS region — used to construct EKS cluster ARNs."
  type        = string
}

variable "github_org" {
  description = "GitHub organisation (or username) that owns the repository."
  type        = string
}

variable "github_repo" {
  description = "GitHub repository name (without the org prefix)."
  type        = string
}

variable "ecr_repo_arns" {
  description = "List of ECR repository ARNs to grant push/pull access to the CI role."
  type        = list(string)
}
