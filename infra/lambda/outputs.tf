output "iam" {
  value = {
    lambda_logging_policy_arn  = aws_iam_policy.lambda_logging.arn
    lambda_logging_policy_id   = aws_iam_policy.lambda_logging.id
    lambda_logging_policy_name = aws_iam_policy.lambda_logging.name
  }
}
