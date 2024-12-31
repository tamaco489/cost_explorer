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

resource "aws_iam_role_policy_attachment" "cost_explorer_logs" {
  policy_arn = data.terraform_remote_state.lambda.outputs.iam.lambda_logging_policy_arn
  role       = aws_iam_role.cost_explorer.name
}
