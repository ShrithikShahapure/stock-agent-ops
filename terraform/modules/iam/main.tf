################################################################################
# IAM module — OIDC providers, CI role, TF Cloud role, EKS cluster + node roles
#
# EKS access entries are created in the eks module (receives ci_role_arn as
# input) to avoid a circular dependency between iam ↔ eks.
################################################################################

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

locals {
  account_id = data.aws_caller_identity.current.account_id
  partition  = data.aws_partition.current.partition
}

# ── GitHub Actions OIDC provider ─────────────────────────────────────────────

data "tls_certificate" "github" {
  url = "https://token.actions.githubusercontent.com/.well-known/openid-configuration"
}

resource "aws_iam_openid_connect_provider" "github" {
  url             = "https://token.actions.githubusercontent.com"
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = [data.tls_certificate.github.certificates[0].sha1_fingerprint]

  tags = { Name = "github-actions-oidc" }
}

# ── Terraform Cloud OIDC provider ────────────────────────────────────────────

data "tls_certificate" "tfc" {
  url = "https://app.terraform.io/.well-known/openid-configuration"
}

resource "aws_iam_openid_connect_provider" "tfc" {
  url             = "https://app.terraform.io"
  client_id_list  = ["aws.workload.identity"]
  thumbprint_list = [data.tls_certificate.tfc.certificates[0].sha1_fingerprint]

  tags = { Name = "terraform-cloud-oidc" }
}

# ── github-actions-ci-role ────────────────────────────────────────────────────

data "aws_iam_policy_document" "github_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [aws_iam_openid_connect_provider.github.arn]
    }

    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:aud"
      values   = ["sts.amazonaws.com"]
    }

    condition {
      test     = "StringLike"
      variable = "token.actions.githubusercontent.com:sub"
      values   = ["repo:${var.github_org}/${var.github_repo}:*"]
    }
  }
}

resource "aws_iam_role" "github_actions_ci" {
  name               = "github-actions-ci-role"
  assume_role_policy = data.aws_iam_policy_document.github_assume.json
  description        = "Assumed by GitHub Actions via OIDC for ECR and EKS operations"
}

data "aws_iam_policy_document" "github_actions_ci" {
  statement {
    sid       = "ECRLogin"
    effect    = "Allow"
    actions   = ["ecr:GetAuthorizationToken"]
    resources = ["*"]
  }

  statement {
    sid    = "ECRPushPull"
    effect = "Allow"
    actions = [
      "ecr:BatchCheckLayerAvailability",
      "ecr:CompleteLayerUpload",
      "ecr:GetDownloadUrlForLayer",
      "ecr:InitiateLayerUpload",
      "ecr:PutImage",
      "ecr:UploadLayerPart",
      "ecr:BatchGetImage",
      "ecr:DescribeImages",
      "ecr:DescribeRepositories",
      "ecr:ListImages",
    ]
    resources = var.ecr_repo_arns
  }

  statement {
    sid    = "EKSDescribe"
    effect = "Allow"
    actions = [
      "eks:DescribeCluster",
      "eks:ListClusters",
    ]
    resources = [
      "arn:${local.partition}:eks:${var.aws_region}:${local.account_id}:cluster/${var.cluster_name}",
    ]
  }

  statement {
    sid    = "TFStateS3"
    effect = "Allow"
    actions = [
      "s3:GetObject",
      "s3:PutObject",
      "s3:DeleteObject",
      "s3:ListBucket",
    ]
    resources = [
      "arn:${local.partition}:s3:::stock-agent-ops-tfstate",
      "arn:${local.partition}:s3:::stock-agent-ops-tfstate/*",
    ]
  }

  statement {
    sid    = "TFStateLock"
    effect = "Allow"
    actions = [
      "dynamodb:GetItem",
      "dynamodb:PutItem",
      "dynamodb:DeleteItem",
    ]
    resources = [
      "arn:${local.partition}:dynamodb:${var.aws_region}:${local.account_id}:table/stock-agent-ops-tflock",
    ]
  }

  # ── Terraform infrastructure management ──────────────────────────────────
  # Required for terraform plan/apply/destroy of VPC, EKS, ECR, and IAM modules.
  # Uses service-level wildcards to avoid chasing individual missing actions.

  statement {
    sid       = "TFStateKMS"
    effect    = "Allow"
    actions   = ["kms:Decrypt", "kms:Encrypt", "kms:GenerateDataKey", "kms:DescribeKey"]
    resources = ["*"]
  }

  statement {
    sid       = "TFManageVPC"
    effect    = "Allow"
    actions   = ["ec2:*"]
    resources = ["*"]
  }

  statement {
    sid       = "TFManageEKS"
    effect    = "Allow"
    actions   = ["eks:*"]
    resources = ["*"]
  }

  statement {
    sid       = "TFManageIAM"
    effect    = "Allow"
    actions   = ["iam:*"]
    resources = ["*"]
  }

  statement {
    sid       = "TFManageECR"
    effect    = "Allow"
    actions   = ["ecr:*"]
    resources = ["*"]
  }

  statement {
    sid       = "TFManageLogs"
    effect    = "Allow"
    actions   = ["logs:*"]
    resources = ["*"]
  }

  statement {
    sid       = "TFManageSTS"
    effect    = "Allow"
    actions   = ["sts:GetCallerIdentity"]
    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "github_actions_ci" {
  name   = "github-actions-ci-policy"
  role   = aws_iam_role.github_actions_ci.id
  policy = data.aws_iam_policy_document.github_actions_ci.json
}

# ── terraform-cloud-role ──────────────────────────────────────────────────────
# NOTE: AdministratorAccess is used here for simplicity.
# TODO: scope down to only the actions Terraform actually needs in production.

data "aws_iam_policy_document" "tfc_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [aws_iam_openid_connect_provider.tfc.arn]
    }

    condition {
      test     = "StringEquals"
      variable = "app.terraform.io:aud"
      values   = ["aws.workload.identity"]
    }

    condition {
      test     = "StringLike"
      variable = "app.terraform.io:sub"
      values   = ["organization:shrithik-shahapure:project:*:workspace:${var.cluster_name}:run_phase:*"]
    }
  }
}

resource "aws_iam_role" "terraform_cloud" {
  name               = "terraform-cloud-role"
  assume_role_policy = data.aws_iam_policy_document.tfc_assume.json
  description        = "Assumed by Terraform Cloud via OIDC dynamic credentials"
}

resource "aws_iam_role_policy_attachment" "terraform_cloud_admin" {
  role       = aws_iam_role.terraform_cloud.name
  policy_arn = "arn:${local.partition}:iam::aws:policy/AdministratorAccess"
}

# ── EKS cluster role ──────────────────────────────────────────────────────────

data "aws_iam_policy_document" "eks_cluster_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["eks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "eks_cluster" {
  name               = "${var.cluster_name}-cluster-role"
  assume_role_policy = data.aws_iam_policy_document.eks_cluster_assume.json
}

resource "aws_iam_role_policy_attachment" "eks_cluster_policy" {
  role       = aws_iam_role.eks_cluster.name
  policy_arn = "arn:${local.partition}:iam::aws:policy/AmazonEKSClusterPolicy"
}

# ── EKS node role ─────────────────────────────────────────────────────────────

data "aws_iam_policy_document" "eks_node_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "eks_node" {
  name               = "${var.cluster_name}-node-role"
  assume_role_policy = data.aws_iam_policy_document.eks_node_assume.json
}

resource "aws_iam_role_policy_attachment" "eks_worker_node" {
  role       = aws_iam_role.eks_node.name
  policy_arn = "arn:${local.partition}:iam::aws:policy/AmazonEKSWorkerNodePolicy"
}

resource "aws_iam_role_policy_attachment" "eks_cni" {
  role       = aws_iam_role.eks_node.name
  policy_arn = "arn:${local.partition}:iam::aws:policy/AmazonEKS_CNI_Policy"
}

resource "aws_iam_role_policy_attachment" "eks_ecr_readonly" {
  role       = aws_iam_role.eks_node.name
  policy_arn = "arn:${local.partition}:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
}

resource "aws_iam_role_policy_attachment" "eks_ebs_csi" {
  role       = aws_iam_role.eks_node.name
  policy_arn = "arn:${local.partition}:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy"
}
