variable "cluster_name" {
  description = "Name of the EKS cluster."
  type        = string
}

variable "kubernetes_version" {
  description = "Kubernetes version for the EKS cluster (e.g. \"1.31\")."
  type        = string
  default     = "1.31"
}

variable "vpc_id" {
  description = "VPC ID — used to create security groups."
  type        = string
}

variable "subnet_ids" {
  description = "All subnet IDs (public + private) attached to the cluster VPC config."
  type        = list(string)
}

variable "private_subnet_ids" {
  description = "Private subnet IDs — worker nodes are placed here."
  type        = list(string)
}

variable "cluster_role_arn" {
  description = "ARN of the IAM role for the EKS control plane."
  type        = string
}

variable "node_role_arn" {
  description = "ARN of the IAM role for EKS worker nodes."
  type        = string
}

variable "ci_role_arn" {
  description = "ARN of the GitHub Actions CI role — granted cluster-admin via EKS access entry."
  type        = string
}

variable "node_instance_type" {
  description = "EC2 instance type for the managed node group."
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
