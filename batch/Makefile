# =================================================================
# common
# =================================================================
VERSION := cost_explorer_v0.0.0


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
	aws ecr get-login-password --region ap-northeast-1 --profile ${AWS_PROFILE} | docker login --username AWS --password-stdin ${AWS_ACCOUNT_ID}.dkr.ecr.ap-northeast-1.amazonaws.com

# make push AWS_ACCOUNT_ID=${AWS_ACCOUNT_ID} AWS_PROFILE=${AWS_PROFILE}
push: auth
	docker build --platform linux/amd64 --no-cache -t $(AWS_ACCOUNT_ID).dkr.ecr.ap-northeast-1.amazonaws.com/dev-cost-explorer:$(VERSION) -f ./build/Dockerfile.cost_explorer .
	docker push ${AWS_ACCOUNT_ID}.dkr.ecr.ap-northeast-1.amazonaws.com/dev-cost-explorer:$(VERSION)

image-list:
	aws ecr list-images --repository-name dev-cost-explorer --region ap-northeast-1 --profile ${AWS_PROFILE} --query 'imageIds[*]' | jq .

# make delete-images AWS_PROFILE=${AWS_PROFILE}
delete-images:
	@aws ecr list-images --repository-name dev-cost-explorer --region ap-northeast-1 --profile ${AWS_PROFILE} --query 'imageIds[*]' --output json > /tmp/images.json
	@cat /tmp/images.json | jq -c '.[]' | while read -r image_id; do \
	  aws ecr batch-delete-image --repository-name dev-cost-explorer --region ap-northeast-1 --profile ${AWS_PROFILE} --image-ids $${image_id}; \
	done | jq .


# =================================================================
# lambda
# =================================================================
.PHONY: deploy
# make deploy AWS_ACCOUNT_ID=${AWS_ACCOUNT_ID} AWS_PROFILE=${AWS_PROFILE}
deploy: push
	aws lambda update-function-code  --profile ${AWS_PROFILE} \
		--function-name dev-cost-explorer \
		--image-uri $(AWS_ACCOUNT_ID).dkr.ecr.ap-northeast-1.amazonaws.com/dev-cost-explorer:$(VERSION) | jq .
