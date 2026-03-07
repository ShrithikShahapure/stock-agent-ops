terraform {
  required_version = ">= 1.7"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }

  backend "s3" {
    bucket         = "stock-agent-ops-tfstate"
    key            = "terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "stock-agent-ops-tflock"
    encrypt        = true
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "stock-agent-ops"
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}

# ── Modules ───────────────────────────────────────────────────────────────────

module "ecr" {
  source      = "./modules/ecr"
  name_prefix = var.cluster_name
}

module "vpc" {
  source             = "./modules/vpc"
  cluster_name       = var.cluster_name
  vpc_cidr           = var.vpc_cidr
  availability_zones = var.availability_zones
}

module "iam" {
  source         = "./modules/iam"
  cluster_name   = var.cluster_name
  aws_region     = var.aws_region
  github_org     = var.github_org
  github_repo    = var.github_repo
  ecr_repo_arns  = [module.ecr.api_repository_arn, module.ecr.frontend_repository_arn]
  enable_sqs     = true
  sqs_queue_arn  = module.sqs.queue_arn
  sqs_dlq_arn    = module.sqs.dlq_arn
  enable_secrets = true
  secrets_arn    = module.secrets.secret_arn
}

module "eks" {
  source             = "./modules/eks"
  cluster_name       = var.cluster_name
  kubernetes_version = var.kubernetes_version
  vpc_id             = module.vpc.vpc_id
  subnet_ids         = concat(module.vpc.public_subnet_ids, module.vpc.private_subnet_ids)
  private_subnet_ids = module.vpc.private_subnet_ids
  cluster_role_arn   = module.iam.eks_cluster_role_arn
  node_role_arn      = module.iam.eks_node_role_arn
  ci_role_arn        = module.iam.github_actions_ci_role_arn
  node_instance_type = var.node_instance_type
  node_min_size      = var.node_min_size
  node_max_size      = var.node_max_size
  node_desired_size  = var.node_desired_size

  # GPU node group
  gpu_enabled       = var.gpu_enabled
  gpu_instance_type = var.gpu_instance_type
  gpu_desired_size  = var.gpu_desired_size
  gpu_max_size      = var.gpu_max_size
  gpu_min_size      = var.gpu_min_size
}

# ── Managed services ─────────────────────────────────────────────────────────

module "elasticache" {
  source                    = "./modules/elasticache"
  cluster_name              = var.cluster_name
  vpc_id                    = module.vpc.vpc_id
  private_subnet_ids        = module.vpc.private_subnet_ids
  node_security_group_id    = module.eks.node_security_group_id
  cluster_security_group_id = module.eks.cluster_security_group_id
}

module "sqs" {
  source       = "./modules/sqs"
  cluster_name = var.cluster_name
}

module "secrets" {
  source            = "./modules/secrets"
  cluster_name      = var.cluster_name
  finnhub_api_key   = var.finnhub_api_key
  slack_webhook_url = var.slack_webhook_url
  alert_email       = var.alert_email
}

module "sagemaker" {
  source       = "./modules/sagemaker"
  cluster_name = var.cluster_name
}
