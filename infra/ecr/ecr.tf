resource "aws_ecr_repository" "cost_explorer" {
  name = "${var.env}-cost-explorer"

  # 既存のタグに対して、後から上書きを可能とする設定
  image_tag_mutability = "MUTABLE"

  # イメージがpushされる度に、自動的にセキュリティスキャンを行う設定を有効にする
  image_scanning_configuration {
    scan_on_push = true
  }

  tags = { Name = "${var.env}-cost-explorer-ecr" }
}

# ライフサイクルポリシーの設定
resource "aws_ecr_lifecycle_policy" "cost_explorer" {
  repository = aws_ecr_repository.cost_explorer.name

  policy = jsonencode(
    {
      "rules" : [
        {
          "rulePriority" : 1,
          "description" : "バージョン付きのイメージを20個保持する、21個目がアップロードされた際には古いものから順に削除されていく",
          "selection" : {
            "tagStatus" : "tagged",
            "tagPrefixList" : ["cost_explorer_v"],
            "countType" : "imageCountMoreThan",
            "countNumber" : 20
          },
          "action" : {
            "type" : "expire"
          }
        },
        {
          "rulePriority" : 2,
          "description" : "タグが設定されていないイメージをアップロードされてから3日後に削除する",
          "selection" : {
            "tagStatus" : "untagged",
            "countType" : "sinceImagePushed",
            "countUnit" : "days",
            "countNumber" : 3
          },
          "action" : {
            "type" : "expire"
          }
        },
        {
          "rulePriority" : 3,
          "description" : "タグが設定されたイメージをアップロードされてから30日後に削除する",
          "selection" : {
            "tagStatus" : "any",
            "countType" : "sinceImagePushed",
            "countUnit" : "days",
            "countNumber" : 30
          },
          "action" : {
            "type" : "expire"
          }
        }
      ]
    }
  )
}