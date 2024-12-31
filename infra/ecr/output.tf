output "cost_explorer" {
  value = {
    arn  = aws_ecr_repository.cost_explorer.arn
    id   = aws_ecr_repository.cost_explorer.id
    name = aws_ecr_repository.cost_explorer.name
    url  = aws_ecr_repository.cost_explorer.repository_url
  }
}
