output "exchange_rates_app_id" {
  value = {
    name = aws_secretsmanager_secret.exchange_rates_app_id.name
    arn  = aws_secretsmanager_secret.exchange_rates_app_id.arn
  }
}
