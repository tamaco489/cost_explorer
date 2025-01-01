resource "aws_lambda_function" "cost_explorer" {
  function_name = local.fqn
  description   = "Cost aggregation batch using AWS Cost Explorer"
  role          = aws_iam_role.cost_explorer.arn
  package_type  = "Image"
  image_uri     = "${data.terraform_remote_state.ecr.outputs.cost_explorer.url}:cost_explorer_v0.0.0"
  timeout       = 20
  memory_size   = 128

  lifecycle {
    ignore_changes = [image_uri]
  }

  environment {
    variables = {
      SERVICE_NAME = "cost-explorer"
      API_ENV      = "dev"
      LOGGING      = "off"
    }
  }

  tags = {
    Name = "${local.fqn}"
  }
}
