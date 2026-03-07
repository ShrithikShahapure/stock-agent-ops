output "execution_role_arn" {
  description = "ARN of the SageMaker execution role."
  value       = aws_iam_role.sagemaker_execution.arn
}

output "bucket_name" {
  description = "Name of the S3 bucket for SageMaker artifacts."
  value       = aws_s3_bucket.sagemaker.id
}

output "bucket_arn" {
  description = "ARN of the S3 bucket for SageMaker artifacts."
  value       = aws_s3_bucket.sagemaker.arn
}
