# Import blocks — adopt all existing AWS resources into Terraform state.
# The previous destroy wiped state but not the actual resources.
# Remove this file after the first successful `terraform apply`.

# ── IAM: OIDC providers ─────────────────────────────────────────────────────
import {
  id = "arn:aws:iam::075223871157:oidc-provider/token.actions.githubusercontent.com"
  to = module.iam.aws_iam_openid_connect_provider.github
}

import {
  id = "arn:aws:iam::075223871157:oidc-provider/app.terraform.io"
  to = module.iam.aws_iam_openid_connect_provider.tfc
}

# ── IAM: Roles + policies ───────────────────────────────────────────────────
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

# ── ECR ──────────────────────────────────────────────────────────────────────
import {
  id = "stock-agent-ops/api"
  to = module.ecr.aws_ecr_repository.api
}

import {
  id = "stock-agent-ops/frontend"
  to = module.ecr.aws_ecr_repository.frontend
}

# ── VPC ──────────────────────────────────────────────────────────────────────
import {
  id = "vpc-0f700414707985add"
  to = module.vpc.aws_vpc.main
}

import {
  id = "igw-0480cf8e5a79ceea7"
  to = module.vpc.aws_internet_gateway.main
}

import {
  id = "eipalloc-0649e759761252892"
  to = module.vpc.aws_eip.nat
}

import {
  id = "nat-041184e5b08264346"
  to = module.vpc.aws_nat_gateway.main
}

# Subnets: public[0]=us-east-1a, public[1]=us-east-1b
import {
  id = "subnet-03b38e98e81c22376"
  to = module.vpc.aws_subnet.public[0]
}

import {
  id = "subnet-0588dfd670cb419cd"
  to = module.vpc.aws_subnet.public[1]
}

# Subnets: private[0]=us-east-1a, private[1]=us-east-1b
import {
  id = "subnet-001d9554cf10f13d3"
  to = module.vpc.aws_subnet.private[0]
}

import {
  id = "subnet-04650111904db7a9c"
  to = module.vpc.aws_subnet.private[1]
}

# Route tables
import {
  id = "rtb-08fbd0725b2a587ab"
  to = module.vpc.aws_route_table.public
}

import {
  id = "rtb-06784f7002afd484c"
  to = module.vpc.aws_route_table.private
}

# Route table associations: public[0]=1a, public[1]=1b
import {
  id = "rtbassoc-0efbd946775fa6f66"
  to = module.vpc.aws_route_table_association.public[0]
}

import {
  id = "rtbassoc-014df23c5d8df5db0"
  to = module.vpc.aws_route_table_association.public[1]
}

# Route table associations: private[0]=1a, private[1]=1b
import {
  id = "rtbassoc-00b185f112d6c32c5"
  to = module.vpc.aws_route_table_association.private[0]
}

import {
  id = "rtbassoc-0f24e30cf1fb474a2"
  to = module.vpc.aws_route_table_association.private[1]
}

# ── EKS ──────────────────────────────────────────────────────────────────────
import {
  id = "sg-0cc844f600292fb3c"
  to = module.eks.aws_security_group.cluster
}

import {
  id = "sg-02a36ea5e2d8baeef"
  to = module.eks.aws_security_group.nodes
}

import {
  id = "stock-agent-ops"
  to = module.eks.aws_eks_cluster.main
}

import {
  id = "stock-agent-ops:stock-agent-ops-nodes"
  to = module.eks.aws_eks_node_group.main
}

# EKS addons
import {
  id = "stock-agent-ops:vpc-cni"
  to = module.eks.aws_eks_addon.vpc_cni
}

import {
  id = "stock-agent-ops:coredns"
  to = module.eks.aws_eks_addon.coredns
}

import {
  id = "stock-agent-ops:kube-proxy"
  to = module.eks.aws_eks_addon.kube_proxy
}

import {
  id = "stock-agent-ops:aws-ebs-csi-driver"
  to = module.eks.aws_eks_addon.ebs_csi_driver
}

# EKS access entries
import {
  id = "stock-agent-ops:arn:aws:iam::075223871157:role/github-actions-ci-role"
  to = module.eks.aws_eks_access_entry.ci
}

import {
  id = "stock-agent-ops#arn:aws:iam::075223871157:role/github-actions-ci-role#arn:aws:eks::aws:cluster-access-policy/AmazonEKSClusterAdminPolicy"
  to = module.eks.aws_eks_access_policy_association.ci_admin
}
