variable "cluster_name" {
  description = "Cluster name — used in resource Name tags and Kubernetes subnet tags."
  type        = string
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC."
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "List of exactly 2 availability zones."
  type        = list(string)
  validation {
    condition     = length(var.availability_zones) == 2
    error_message = "Exactly 2 availability zones are required."
  }
}
