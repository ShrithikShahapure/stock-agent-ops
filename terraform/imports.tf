# Import blocks — adopt the OIDC provider and CI role created by the
# bootstrap script (scripts/bootstrap-aws.sh).
# Remove these after the first successful `terraform apply`.

import {
  id = "arn:aws:iam::075223871157:oidc-provider/token.actions.githubusercontent.com"
  to = module.iam.aws_iam_openid_connect_provider.github
}

import {
  id = "github-actions-ci-role"
  to = module.iam.aws_iam_role.github_actions_ci
}

import {
  id = "github-actions-ci-role:github-actions-ci-policy"
  to = module.iam.aws_iam_role_policy.github_actions_ci
}
