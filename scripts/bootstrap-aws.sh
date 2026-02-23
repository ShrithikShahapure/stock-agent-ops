#!/usr/bin/env bash
# bootstrap-aws.sh — Idempotent setup of AWS resources that must exist BEFORE
# Terraform or GitHub Actions pipelines can run:
#   1. S3 bucket   (Terraform state)
#   2. DynamoDB table (Terraform state lock)
#   3. GitHub OIDC provider
#   4. github-actions-ci-role (OIDC trust + broad infra permissions)
#
# Prerequisites:
#   - AWS CLI v2 configured with credentials that have IAM + S3 + DynamoDB admin
#   - jq installed
#
# Usage:
#   ./scripts/bootstrap-aws.sh                 # uses defaults
#   AWS_REGION=us-west-2 ./scripts/bootstrap-aws.sh  # override region

set -euo pipefail

# ── Configuration (override via env) ──────────────────────────────────────────
REGION="${AWS_REGION:-us-east-1}"
STATE_BUCKET="${TF_STATE_BUCKET:-stock-agent-ops-tfstate}"
LOCK_TABLE="${TF_LOCK_TABLE:-stock-agent-ops-tflock}"
ROLE_NAME="${CI_ROLE_NAME:-github-actions-ci-role}"
GITHUB_ORG="${GITHUB_ORG:-ShrithikShahapure}"
GITHUB_REPO="${GITHUB_REPO:-stock-agent-ops}"
OIDC_URL="https://token.actions.githubusercontent.com"
OIDC_AUDIENCE="sts.amazonaws.com"

ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
echo "AWS Account: ${ACCOUNT_ID}"
echo "Region:      ${REGION}"
echo ""

# ── 1. S3 state bucket ───────────────────────────────────────────────────────
echo "==> S3 state bucket: ${STATE_BUCKET}"
if aws s3api head-bucket --bucket "${STATE_BUCKET}" 2>/dev/null; then
  echo "    Already exists."
else
  echo "    Creating..."
  if [ "${REGION}" = "us-east-1" ]; then
    aws s3api create-bucket --bucket "${STATE_BUCKET}" --region "${REGION}"
  else
    aws s3api create-bucket --bucket "${STATE_BUCKET}" --region "${REGION}" \
      --create-bucket-configuration LocationConstraint="${REGION}"
  fi
  aws s3api put-bucket-versioning --bucket "${STATE_BUCKET}" \
    --versioning-configuration Status=Enabled
  aws s3api put-bucket-encryption --bucket "${STATE_BUCKET}" \
    --server-side-encryption-configuration '{
      "Rules": [{"ApplyServerSideEncryptionByDefault": {"SSEAlgorithm": "AES256"}}]
    }'
  aws s3api put-public-access-block --bucket "${STATE_BUCKET}" \
    --public-access-block-configuration \
      BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true
  echo "    Created."
fi

# ── 2. DynamoDB lock table ────────────────────────────────────────────────────
echo "==> DynamoDB lock table: ${LOCK_TABLE}"
if aws dynamodb describe-table --table-name "${LOCK_TABLE}" --region "${REGION}" >/dev/null 2>&1; then
  echo "    Already exists."
else
  echo "    Creating..."
  aws dynamodb create-table \
    --table-name "${LOCK_TABLE}" \
    --attribute-definitions AttributeName=LockID,AttributeType=S \
    --key-schema AttributeName=LockID,KeyType=HASH \
    --billing-mode PAY_PER_REQUEST \
    --region "${REGION}"
  aws dynamodb wait table-exists --table-name "${LOCK_TABLE}" --region "${REGION}"
  echo "    Created."
fi

# ── 3. GitHub OIDC provider ──────────────────────────────────────────────────
echo "==> GitHub OIDC provider"
OIDC_ARN=$(aws iam list-open-id-connect-providers --query \
  "OpenIDConnectProviderList[?ends_with(Arn, '/token.actions.githubusercontent.com')].Arn | [0]" \
  --output text 2>/dev/null || true)

if [ -n "${OIDC_ARN}" ] && [ "${OIDC_ARN}" != "None" ]; then
  echo "    Already exists: ${OIDC_ARN}"
else
  echo "    Creating..."
  # Fetch the current TLS thumbprint
  THUMBPRINT=$(openssl s_client -servername token.actions.githubusercontent.com \
    -connect token.actions.githubusercontent.com:443 </dev/null 2>/dev/null \
    | openssl x509 -fingerprint -noout -sha1 2>/dev/null \
    | sed 's/sha1 Fingerprint=//;s/://g' \
    | tr '[:upper:]' '[:lower:]')

  # Fallback to well-known thumbprint if openssl fails
  if [ -z "${THUMBPRINT}" ]; then
    THUMBPRINT="6938fd4d98bab03faadb97b34396831e3780aea1"
  fi

  OIDC_ARN=$(aws iam create-open-id-connect-provider \
    --url "${OIDC_URL}" \
    --client-id-list "${OIDC_AUDIENCE}" \
    --thumbprint-list "${THUMBPRINT}" \
    --tags Key=Name,Value=github-actions-oidc Key=ManagedBy,Value=bootstrap \
    --query OpenIDConnectProviderArn --output text)
  echo "    Created: ${OIDC_ARN}"
fi

# ── 4. CI role ────────────────────────────────────────────────────────────────
echo "==> CI role: ${ROLE_NAME}"
if aws iam get-role --role-name "${ROLE_NAME}" >/dev/null 2>&1; then
  echo "    Already exists. Updating trust policy..."
else
  echo "    Creating..."
fi

# Trust policy — allow GitHub Actions OIDC for this org/repo on any branch/event
TRUST_POLICY=$(cat <<TRUST
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": { "Federated": "${OIDC_ARN}" },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "${OIDC_AUDIENCE}"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:${GITHUB_ORG}/${GITHUB_REPO}:*"
        }
      }
    }
  ]
}
TRUST
)

if aws iam get-role --role-name "${ROLE_NAME}" >/dev/null 2>&1; then
  aws iam update-assume-role-policy --role-name "${ROLE_NAME}" \
    --policy-document "${TRUST_POLICY}"
else
  aws iam create-role --role-name "${ROLE_NAME}" \
    --assume-role-policy-document "${TRUST_POLICY}" \
    --description "Assumed by GitHub Actions via OIDC for CI/CD operations" \
    --tags Key=ManagedBy,Value=bootstrap
fi

# Inline policy — matches the Terraform IAM module policy
POLICY=$(cat <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "ECRLogin",
      "Effect": "Allow",
      "Action": "ecr:GetAuthorizationToken",
      "Resource": "*"
    },
    {
      "Sid": "ECRPushPull",
      "Effect": "Allow",
      "Action": [
        "ecr:BatchCheckLayerAvailability", "ecr:CompleteLayerUpload",
        "ecr:GetDownloadUrlForLayer", "ecr:InitiateLayerUpload",
        "ecr:PutImage", "ecr:UploadLayerPart", "ecr:BatchGetImage",
        "ecr:DescribeImages", "ecr:DescribeRepositories", "ecr:ListImages"
      ],
      "Resource": [
        "arn:aws:ecr:${REGION}:${ACCOUNT_ID}:repository/stock-agent-ops/*"
      ]
    },
    {
      "Sid": "TFStateS3",
      "Effect": "Allow",
      "Action": ["s3:GetObject", "s3:PutObject", "s3:DeleteObject", "s3:ListBucket"],
      "Resource": [
        "arn:aws:s3:::${STATE_BUCKET}",
        "arn:aws:s3:::${STATE_BUCKET}/*"
      ]
    },
    {
      "Sid": "TFStateLock",
      "Effect": "Allow",
      "Action": ["dynamodb:GetItem", "dynamodb:PutItem", "dynamodb:DeleteItem"],
      "Resource": "arn:aws:dynamodb:${REGION}:${ACCOUNT_ID}:table/${LOCK_TABLE}"
    },
    {
      "Sid": "TFStateKMS",
      "Effect": "Allow",
      "Action": ["kms:Decrypt", "kms:Encrypt", "kms:GenerateDataKey", "kms:DescribeKey"],
      "Resource": "*"
    },
    { "Sid": "TFManageVPC",  "Effect": "Allow", "Action": "ec2:*",  "Resource": "*" },
    { "Sid": "TFManageEKS",  "Effect": "Allow", "Action": "eks:*",  "Resource": "*" },
    { "Sid": "TFManageIAM",  "Effect": "Allow", "Action": "iam:*",  "Resource": "*" },
    { "Sid": "TFManageECR",  "Effect": "Allow", "Action": "ecr:*",  "Resource": "*" },
    { "Sid": "TFManageLogs", "Effect": "Allow", "Action": "logs:*", "Resource": "*" },
    { "Sid": "TFManageSTS",  "Effect": "Allow", "Action": "sts:GetCallerIdentity", "Resource": "*" }
  ]
}
POLICY
)

aws iam put-role-policy --role-name "${ROLE_NAME}" \
  --policy-name "github-actions-ci-policy" \
  --policy-document "${POLICY}"
echo "    Policy applied."

# ── Summary ──────────────────────────────────────────────────────────────────
echo ""
echo "Bootstrap complete."
echo "  S3 bucket:     ${STATE_BUCKET}"
echo "  DynamoDB table: ${LOCK_TABLE}"
echo "  OIDC provider: ${OIDC_ARN}"
echo "  CI role ARN:   arn:aws:iam::${ACCOUNT_ID}:role/${ROLE_NAME}"
echo ""
echo "Next steps:"
echo "  1. Set GitHub secret AWS_ACCOUNT_ID = ${ACCOUNT_ID}"
echo "  2. Run: terraform -chdir=terraform init"
echo "  3. Run: terraform -chdir=terraform import aws_iam_openid_connect_provider.github ${OIDC_ARN}"
echo "  4. Run: terraform -chdir=terraform import aws_iam_role.github_actions_ci ${ROLE_NAME}"
echo "     (or let Terraform adopt them on next apply)"
