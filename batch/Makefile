# =================================================================
# common
# =================================================================
ENV           ?= dev
VERSION       := cost_explorer_v0.0.0
AWS_REGION    := ap-northeast-1
IMAGE_NAME    := $(ENV)-cost-explorer

.PHONY: help
help: ## ヘルプ
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'


# =================================================================
# ci
# =================================================================
.PHONY: lint deps build run
lint: ## golangci-lintによる静的解析
	golangci-lint run --timeout 3m

deps: ## go.mod の依存性解決
	go mod download && go mod tidy

build: deps ## main.goのビルド
	go build -o ./build/cost_explorer ./cmd/main.go

run: build ## ビルドファイルを実行 (ci向け)
	./build/cost_explorer


# =================================================================
# ecr
# =================================================================
.PHONY: auth push image-list delete-images
auth: ## ecrのセッションを取得
	aws ecr get-login-password --region $(AWS_REGION) --profile ${AWS_PROFILE} | docker login --username AWS --password-stdin ${AWS_ACCOUNT_ID}.dkr.ecr.$(AWS_REGION).amazonaws.com

# make push AWS_ACCOUNT_ID=${AWS_ACCOUNT_ID} AWS_PROFILE=${AWS_PROFILE}
push: auth ## localで生成したコンテナイメージをecrへpush
	docker build --platform linux/amd64 --no-cache -t $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/$(IMAGE_NAME):$(VERSION) -f ./build/Dockerfile.cost_explorer .
	docker push ${AWS_ACCOUNT_ID}.dkr.ecr.$(AWS_REGION).amazonaws.com/$(IMAGE_NAME):$(VERSION)

image-list: ## ecrで管理しているイメージのリストを取得
	aws ecr list-images --repository-name $(IMAGE_NAME) --region $(AWS_REGION) --profile ${AWS_PROFILE} --query 'imageIds[*]' | jq .

# make delete-images AWS_PROFILE=${AWS_PROFILE}
delete-images: ## ecrで管理しているイメージを全て削除
	@aws ecr list-images --repository-name $(IMAGE_NAME) --region $(AWS_REGION) --profile ${AWS_PROFILE} --query 'imageIds[*]' --output json > /tmp/images.json
	@cat /tmp/images.json | jq -c '.[]' | while read -r image_id; do \
	  aws ecr batch-delete-image --repository-name $(IMAGE_NAME) --region $(AWS_REGION) --profile ${AWS_PROFILE} --image-ids $${image_id}; \
	done | jq .


# =================================================================
# lambda
# =================================================================
FUNCTION_NAME := $(ENV)-cost-explorer
OUTPUT_JSON   := payload/result/output.json
ENCODED_PAYLOAD_DAILY := $(shell echo -n '{"type": "dailyCostReport"}' | base64)
ENCODED_PAYLOAD_WEEKLY := $(shell echo -n '{"type": "weeklyCostReport"}' | base64)
ENCODED_PAYLOAD_MONTHLY := $(shell echo -n '{"type": "monthlyCostReport"}' | base64)

.PHONY: deploy invoke-daily invoke-weekly invoke-monthly

# make deploy AWS_ACCOUNT_ID=${AWS_ACCOUNT_ID} AWS_PROFILE=${AWS_PROFILE}
deploy: push ## lambdaが参照しているecrコンテナをアップデート
	@echo "Deploy Lambda"
	aws lambda update-function-code \
		--profile $(AWS_PROFILE) \
		--function-name $(FUNCTION_NAME) \
		--image-uri $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/$(IMAGE_NAME):$(VERSION) | jq .

# $ make invoke-daily AWS_PROFILE=${AWS_PROFILE}
invoke-daily: ## 日次レポート送信処理を実行
	@echo "Invoking Lambda with event type: dailyCostReport"
	aws lambda invoke \
		--profile $(AWS_PROFILE) \
		--function-name $(FUNCTION_NAME) \
		--payload "$(ENCODED_PAYLOAD_DAILY)" \
		$(OUTPUT_JSON) | jq .

invoke-weekly: ## 週次レポート送信処理を実行
	@echo "Invoking Lambda with event type: weeklyCostReport"
	aws lambda invoke \
		--profile $(AWS_PROFILE) \
		--function-name $(FUNCTION_NAME) \
		--payload "$(ENCODED_PAYLOAD_WEEKLY)" \
		$(OUTPUT_JSON) | jq .

invoke-monthly: ## 月次レポート送信処理を実行
	@echo "Invoking Lambda with event type: monthlyCostReport"
	aws lambda invoke \
		--profile $(AWS_PROFILE) \
		--function-name $(FUNCTION_NAME) \
		--payload "$(ENCODED_PAYLOAD_MONTHLY)" \
		$(OUTPUT_JSON) | jq .


# =================================================================
# secret manager
# =================================================================
# $ make get-secret AWS_PROFILE=${AWS_PROFILE}
.PHONY: get-secret
get-secret: ## secret manager に登録した secret を一括で取得
	@echo "Batch get secret value"
	aws secretsmanager batch-get-secret-value \
		--profile $(AWS_PROFILE) \
		--secret-id-list "cost-explorer/$(ENV)/slack/config" "cost-explorer/$(ENV)/exchange-rates/app-id" | jq .
