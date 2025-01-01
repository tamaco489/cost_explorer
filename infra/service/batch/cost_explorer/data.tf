data "terraform_remote_state" "ecr" {
  backend = "s3"
  config = {
    bucket = "dev-cost-explorer-tfstate"
    key    = "ecr/terraform.tfstate"
  }
}

data "terraform_remote_state" "lambda" {
  backend = "s3"
  config = {
    bucket = "dev-cost-explorer-tfstate"
    key    = "lambda/terraform.tfstate"
  }
}

data "aws_secretsmanager_secret" "slack_config" {
  name = "${var.product}/${var.env}/slack/config"
}

data "aws_secretsmanager_secret" "exchange_rates_app_id" {
  name = "${var.product}/${var.env}/exchange-rates/app-id"
}
