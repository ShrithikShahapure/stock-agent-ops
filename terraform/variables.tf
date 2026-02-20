variable "aws_region" {
  description = "AWS region to deploy all resources into."
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Deployment environment label (e.g. prod, staging)."
  type        = string
  default     = "prod"
}

variable "cluster_name" {
  description = "Name of the EKS cluster."
  type        = string
  default     = "stock-agent-ops"
}

variable "kubernetes_version" {
  description = "Kubernetes version for the EKS cluster."
  type        = string
  default     = "1.31"
}

variable "node_instance_type" {
  description = "EC2 instance type for EKS managed node group."
  type        = string
  default     = "m5.xlarge"
}

variable "node_min_size" {
  description = "Minimum number of worker nodes."
  type        = number
  default     = 2
}

variable "node_max_size" {
  description = "Maximum number of worker nodes."
  type        = number
  default     = 4
}

variable "node_desired_size" {
  description = "Desired number of worker nodes."
  type        = number
  default     = 2
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC."
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "List of AZs to use (2 required)."
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b"]
}

variable "github_org" {
  description = "GitHub organisation or user that owns the repository."
  type        = string
  default     = "ShrithikShahapure"
}

variable "github_repo" {
  description = "GitHub repository name (without org prefix)."
  type        = string
  default     = "stock-agent-ops"
}

# ── Sensitive variables — set in Terraform Cloud workspace ──────────────────
variable "finnhub_api_key" {
  description = "Finnhub API key (sensitive — set in TFC workspace variables)."
  type        = string
  sensitive   = true
  default     = ""
}

variable "slack_webhook_url" {
  description = "Slack incoming webhook URL (sensitive — set in TFC workspace variables)."
  type        = string
  sensitive   = true
  default     = ""
}

variable "alert_email" {
  description = "Email address for CloudWatch / SNS alerts."
  type        = string
  sensitive   = true
  default     = ""
}
