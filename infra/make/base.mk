DEFAULT_ENV := dev

ifeq ($(ENV),stg)
	AWS_ACCESS_KEY_ID := STG_ACCESS_KEY_ID
	AWS_SECRET_ACCESS_KEY := STG_SECRET_ACCESS_KEY
	AWS_REGION := ap-northeast-1
else ifeq ($(ENV),prd)
	AWS_ACCESS_KEY_ID := PRD_ACCESS_KEY_ID
	AWS_SECRET_ACCESS_KEY := PRD_SECRET_ACCESS_KEY
	AWS_REGION := ap-northeast-1
else
	AWS_ACCESS_KEY_ID := $(shell echo $$AWS_ACCESS_KEY_ID)
	AWS_SECRET_ACCESS_KEY := $(shell echo $$AWS_SECRET_ACCESS_KEY)
	AWS_REGION := ap-northeast-1
endif

.PHONY: fmt init list show plan apply destroy

fmt:
	terraform fmt

init:
	@export AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) && \
	export AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) && \
	export AWS_REGION=$(AWS_REGION) && \
	terraform init -reconfigure

list:
	@export AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) && \
	export AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) && \
	export AWS_REGION=$(AWS_REGION) && \
	terraform state list

# e.g. make show AWS_RESOURCE=aws_acm_certificate.nautilus
show:
	@export AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) && \
	export AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) && \
	export AWS_REGION=$(AWS_REGION) && \
	terraform state show $(AWS_RESOURCE)

plan:
	@export AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) && \
	export AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) && \
	export AWS_REGION=$(AWS_REGION) && \
	terraform plan

apply:
	@export AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) && \
	export AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) && \
	export AWS_REGION=$(AWS_REGION) && \
	terraform apply

destroy:
	@export AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) && \
	export AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) && \
	export AWS_REGION=$(AWS_REGION) && \
	terraform destroy
