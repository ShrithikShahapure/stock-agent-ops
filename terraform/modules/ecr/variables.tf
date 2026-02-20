variable "name_prefix" {
  description = "Prefix for ECR repository names. Repos are created as <prefix>/api and <prefix>/frontend."
  type        = string
  default     = "stock-agent-ops"
}
