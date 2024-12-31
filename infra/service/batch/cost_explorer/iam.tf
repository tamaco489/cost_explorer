# =================================================================
# lambda function
# =================================================================
data "aws_iam_policy_document" "lambda_execution_assume_role" {
  statement {
    effect = "Allow"
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "cost_explorer" {
  name               = "${local.fqn}-iam-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_execution_assume_role.json
  tags = {
    Name = "${local.fqn}-iam-role"
  }
}

resource "aws_iam_role_policy" "cost_explorer" {
  name = "${local.fqn}-iam-role-policy"
  role = aws_iam_role.cost_explorer.name
  policy = data.aws_iam_policy_document.cost_explorer.json
}

data "aws_iam_policy_document" "cost_explorer" {
  statement {
    effect  = "Allow"
    actions = ["secretsmanager:GetSecretValue"]
    resources = [
      data.aws_secretsmanager_secret.slack_config.arn,
    ]
  }
  statement {
    effect  = "Allow"
    actions = [
      "ce:GetCostAndUsage"
    ]
    // NOTE: 検証が済んだら権限制限しよう
    resources = ["*"]
  }
}

resource "aws_iam_role_policy_attachment" "cost_explorer_logs" {
  policy_arn = data.terraform_remote_state.lambda.outputs.iam.lambda_logging_policy_arn
  role       = aws_iam_role.cost_explorer.name
}


# =================================================================
# event bridge scheduler
# =================================================================
data "aws_iam_policy_document" "eventbridge_scheduler_assume_policy" {
  statement {
    effect = "Allow"
    actions = [
      "sts:AssumeRole",
    ]
    principals {
      type = "Service"
      identifiers = [
        "scheduler.amazonaws.com",
      ]
    }
  }
}

resource "aws_iam_role" "evnetbridge_scheduler" {
  name               = "${local.fqn}-evnetbridge-scheduler-role"
  assume_role_policy = data.aws_iam_policy_document.eventbridge_scheduler_assume_policy.json
}

resource "aws_iam_role_policy" "eventbridge_scheduler" {
  name = "${local.fqn}-eventbridge-scheduler-role-policy"
  role = aws_iam_role.evnetbridge_scheduler.name
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "lambda:InvokeFunction",
        ]
        Effect   = "Allow"
        Resource = aws_lambda_function.cost_explorer.arn
      },
    ]
  })
}