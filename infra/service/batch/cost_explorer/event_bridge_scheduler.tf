resource "aws_scheduler_schedule" "daily_report" {
  name        = "${local.fqn}-daily-report"
  description = "毎日AM08:15に日次集計レポートを送信"
  group_name  = "default"

  flexible_time_window {
    mode = "OFF"
  }

  state = "ENABLED"

  schedule_expression          = "cron(15 8 * * ? *)"
  schedule_expression_timezone = "Asia/Tokyo"

  target {
    arn      = aws_lambda_function.cost_explorer.arn
    role_arn = aws_iam_role.evnetbridge_scheduler.arn
    retry_policy {
      maximum_retry_attempts = 0
    }

    input = jsonencode({
      "type" = "dailyCostReport"
    })
  }
}

resource "aws_scheduler_schedule" "weekly_report" {
  name        = "${local.fqn}-weekly-report"
  description = "毎週月曜AM08:45に週次集計レポートを送信"
  group_name  = "default"

  flexible_time_window {
    mode = "OFF"
  }

  state = "ENABLED"

  schedule_expression          = "cron(45 8 ? * 2 *)"
  schedule_expression_timezone = "Asia/Tokyo"

  target {
    arn      = aws_lambda_function.cost_explorer.arn
    role_arn = aws_iam_role.evnetbridge_scheduler.arn
    retry_policy {
      maximum_retry_attempts = 0
    }

    input = jsonencode({
      "type" = "weeklyCostReport"
    })
  }
}
