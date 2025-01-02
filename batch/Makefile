# =================================================================
# common
# =================================================================
ENV           ?= dev
VERSION       := cost_explorer_v0.0.0
AWS_REGION    := ap-northeast-1
IMAGE_NAME    := $(ENV)-cost-explorer


# =================================================================
# ci
# =================================================================
.PHONY: lint deps build run
lint:
	golangci-lint run --timeout 3m

deps:
	go mod download && go mod tidy

build: deps
	go build -o ./build/cost_explorer ./cmd/main.go

run: build
	./build/cost_explorer


# =================================================================
# ecr
# =================================================================
.PHONY: auth push image-list delete-images
auth:
	aws ecr get-login-password --region $(AWS_REGION) --profile ${AWS_PROFILE} | docker login --username AWS --password-stdin ${AWS_ACCOUNT_ID}.dkr.ecr.$(AWS_REGION).amazonaws.com

# make push AWS_ACCOUNT_ID=${AWS_ACCOUNT_ID} AWS_PROFILE=${AWS_PROFILE}
push: auth
	docker build --platform linux/amd64 --no-cache -t $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/$(IMAGE_NAME):$(VERSION) -f ./build/Dockerfile.cost_explorer .
	docker push ${AWS_ACCOUNT_ID}.dkr.ecr.$(AWS_REGION).amazonaws.com/$(IMAGE_NAME):$(VERSION)

image-list:
	aws ecr list-images --repository-name $(IMAGE_NAME) --region $(AWS_REGION) --profile ${AWS_PROFILE} --query 'imageIds[*]' | jq .

# make delete-images AWS_PROFILE=${AWS_PROFILE}
delete-images:
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

# $ make get-secret AWS_PROFILE=${AWS_PROFILE}
.PHONY: deploy invoke-daily
deploy: push
	aws lambda update-function-code  --profile $(AWS_PROFILE) \
		--function-name $(FUNCTION_NAME) \
		--image-uri $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/$(IMAGE_NAME):$(VERSION) | jq .

invoke-daily:
	@echo "Invoking Lambda with event type: dailyCostReport"
	aws lambda invoke \
		--profile $(AWS_PROFILE) \
		--function-name $(FUNCTION_NAME) \
		--payload "$(ENCODED_PAYLOAD_DAILY)" \
		$(OUTPUT_JSON) | jq .

invoke-weekly:
	@echo "Invoking Lambda with event type: weeklyCostReport"
	aws lambda invoke \
		--profile $(AWS_PROFILE) \
		--function-name $(FUNCTION_NAME) \
		--payload "$(ENCODED_PAYLOAD_WEEKLY)" \
		$(OUTPUT_JSON) | jq .


# =================================================================
# secret manager
# =================================================================
# $ make get-secret AWS_PROFILE=${AWS_PROFILE}
.PHONY: get-secret
get-secret:
	aws secretsmanager batch-get-secret-value \
		--profile $(AWS_PROFILE) \
		--secret-id-list "cost-explorer/$(ENV)/slack/config" "cost-explorer/$(ENV)/exchange-rates/app-id" | jq .
