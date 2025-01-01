resource "aws_secretsmanager_secret" "exchange_rates_app_id" {
  name        = "${var.product}/${var.env}/exchange-rates/app-id"
  description = "manage confidential information on exchange rates"
}

# NOTE: 一度リソースのみダミー文字列で作成して、AWSマネジメントコンソール上で直接編集する
resource "aws_secretsmanager_secret_version" "exchange_rates_app_id" {
  secret_id     = aws_secretsmanager_secret.exchange_rates_app_id.id
  secret_string = var.exchange_rates_app_id
}
