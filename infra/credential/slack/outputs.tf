output "slack_config" {
  value = {
    name = aws_secretsmanager_secret.slack_config.name
    arn  = aws_secretsmanager_secret.slack_config.arn
  }
}
