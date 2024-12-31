resource "aws_secretsmanager_secret" "slack_config" {
  name        = "${var.product}/${var.env}/slack"
  description = "manage confidential information on slack"
}

# NOTE: 一度リソースのみダミー文字列で作成して、AWSマネジメントコンソール上で直接編集する
resource "aws_secretsmanager_secret_version" "slack_config_secret" {
  secret_id     = aws_secretsmanager_secret.slack_config.id
  secret_string = jsonencode(var.slack_config)
}
