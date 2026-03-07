################################################################################
# SageMaker module — S3 bucket for training artifacts, IAM execution role
################################################################################

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

locals {
  account_id = data.aws_caller_identity.current.account_id
  partition  = data.aws_partition.current.partition
  region     = data.aws_region.current.name
}

# ── S3 bucket for SageMaker data, models, and evaluation artifacts ───────────

resource "aws_s3_bucket" "sagemaker" {
  bucket        = "${var.cluster_name}-sagemaker-${local.account_id}"
  force_destroy = true

  tags = {
    Name = "${var.cluster_name}-sagemaker"
  }
}

resource "aws_s3_bucket_versioning" "sagemaker" {
  bucket = aws_s3_bucket.sagemaker.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "sagemaker" {
  bucket = aws_s3_bucket.sagemaker.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "sagemaker" {
  bucket                  = aws_s3_bucket.sagemaker.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_lifecycle_configuration" "sagemaker" {
  bucket = aws_s3_bucket.sagemaker.id

  rule {
    id     = "expire-old-training-data"
    status = "Enabled"
    filter {
      prefix = "sagemaker/data/"
    }
    expiration {
      days = 30
    }
  }

  rule {
    id     = "expire-old-evaluations"
    status = "Enabled"
    filter {
      prefix = "sagemaker/evaluation/"
    }
    expiration {
      days = 90
    }
  }
}

# ── SageMaker execution role ────────────────────────────────────────────────

data "aws_iam_policy_document" "sagemaker_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "sagemaker_execution" {
  name               = "${var.cluster_name}-sagemaker-execution"
  assume_role_policy = data.aws_iam_policy_document.sagemaker_assume.json
  description        = "SageMaker execution role for training pipeline"

  tags = {
    Name = "${var.cluster_name}-sagemaker-execution"
  }
}

# S3 access for training data and model artifacts
data "aws_iam_policy_document" "sagemaker_s3" {
  statement {
    sid    = "S3BucketAccess"
    effect = "Allow"
    actions = [
      "s3:GetObject",
      "s3:PutObject",
      "s3:DeleteObject",
      "s3:ListBucket",
      "s3:GetBucketLocation",
    ]
    resources = [
      aws_s3_bucket.sagemaker.arn,
      "${aws_s3_bucket.sagemaker.arn}/*",
    ]
  }
}

resource "aws_iam_role_policy" "sagemaker_s3" {
  name   = "${var.cluster_name}-sagemaker-s3"
  role   = aws_iam_role.sagemaker_execution.id
  policy = data.aws_iam_policy_document.sagemaker_s3.json
}

# ECR access for pulling SageMaker framework containers
resource "aws_iam_role_policy_attachment" "sagemaker_ecr" {
  role       = aws_iam_role.sagemaker_execution.name
  policy_arn = "arn:${local.partition}:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
}

# CloudWatch Logs for training job logs
data "aws_iam_policy_document" "sagemaker_logs" {
  statement {
    sid    = "CloudWatchLogs"
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:DescribeLogStreams",
    ]
    resources = [
      "arn:${local.partition}:logs:${local.region}:${local.account_id}:log-group:/aws/sagemaker/*",
    ]
  }
}

resource "aws_iam_role_policy" "sagemaker_logs" {
  name   = "${var.cluster_name}-sagemaker-logs"
  role   = aws_iam_role.sagemaker_execution.id
  policy = data.aws_iam_policy_document.sagemaker_logs.json
}

# SageMaker full access for pipeline operations
resource "aws_iam_role_policy_attachment" "sagemaker_full" {
  role       = aws_iam_role.sagemaker_execution.name
  policy_arn = "arn:${local.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}

# PassRole so SageMaker can pass this role to training/processing jobs
data "aws_iam_policy_document" "sagemaker_passrole" {
  statement {
    sid     = "PassRole"
    effect  = "Allow"
    actions = ["iam:PassRole"]
    resources = [
      aws_iam_role.sagemaker_execution.arn,
    ]
    condition {
      test     = "StringEquals"
      variable = "iam:PassedToService"
      values   = ["sagemaker.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy" "sagemaker_passrole" {
  name   = "${var.cluster_name}-sagemaker-passrole"
  role   = aws_iam_role.sagemaker_execution.id
  policy = data.aws_iam_policy_document.sagemaker_passrole.json
}
