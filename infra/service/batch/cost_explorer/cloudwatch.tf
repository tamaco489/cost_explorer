resource "aws_cloudwatch_log_group" "cost_explorer" {
  name              = "/aws/lambda/${aws_lambda_function.cost_explorer.function_name}"
  retention_in_days = 3

  tags = {
    Name = "${local.fqn}-cloudwatch-logs"
  }
}
