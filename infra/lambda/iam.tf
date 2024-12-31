data "aws_iam_policy_document" "lambda_logging" {
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["arn:aws:logs:*:*:*"]
  }
}

resource "aws_iam_policy" "lambda_logging" {
  name        = "cost-explorer-lambda-common-logging-iam-policy"
  path        = "/"
  description = "IAM policy for Lambda logging"
  policy      = data.aws_iam_policy_document.lambda_logging.json
}
