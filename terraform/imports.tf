# Import blocks — tells Terraform to adopt the resources bootstrapped manually
# via AWS CLI rather than creating them from scratch.
# These can be removed after the first successful `terraform apply`.

import {
  id = "arn:aws:iam::075223871157:oidc-provider/token.actions.githubusercontent.com"
  to = module.iam.aws_iam_openid_connect_provider.github
}

import {
  id = "arn:aws:iam::075223871157:oidc-provider/app.terraform.io"
  to = module.iam.aws_iam_openid_connect_provider.tfc
}

import {
  id = "github-actions-ci-role"
  to = module.iam.aws_iam_role.github_actions_ci
}

import {
  id = "github-actions-ci-role:github-actions-ci-policy"
  to = module.iam.aws_iam_role_policy.github_actions_ci
}

import {
  id = "terraform-cloud-role"
  to = module.iam.aws_iam_role.terraform_cloud
}

import {
  id = "terraform-cloud-role/arn:aws:iam::aws:policy/AdministratorAccess"
  to = module.iam.aws_iam_role_policy_attachment.terraform_cloud_admin
}

import {
  id = "stock-agent-ops-cluster-role"
  to = module.iam.aws_iam_role.eks_cluster
}

import {
  id = "stock-agent-ops-cluster-role/arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  to = module.iam.aws_iam_role_policy_attachment.eks_cluster_policy
}

import {
  id = "stock-agent-ops-node-role"
  to = module.iam.aws_iam_role.eks_node
}

import {
  id = "stock-agent-ops-node-role/arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  to = module.iam.aws_iam_role_policy_attachment.eks_worker_node
}

import {
  id = "stock-agent-ops-node-role/arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  to = module.iam.aws_iam_role_policy_attachment.eks_cni
}

import {
  id = "stock-agent-ops-node-role/arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  to = module.iam.aws_iam_role_policy_attachment.eks_ecr_readonly
}

import {
  id = "stock-agent-ops-node-role/arn:aws:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy"
  to = module.iam.aws_iam_role_policy_attachment.eks_ebs_csi
}

import {
  id = "stock-agent-ops/api"
  to = module.ecr.aws_ecr_repository.api
}

import {
  id = "stock-agent-ops/frontend"
  to = module.ecr.aws_ecr_repository.frontend
}
