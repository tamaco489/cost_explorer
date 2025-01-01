variable "env" {
  description = "environment name"
  type        = string
  default     = "dev"
}

variable "product" {
  description = "product name"
  type        = string
  default     = "cost-explorer"
}

variable "region" {
  description = "region name"
  type        = string
  default     = "ap-northeast-1"
}

locals {
  fqn = "${var.env}-${var.product}"
}

variable "exchange_rates_app_id" {
  description = "value"
  type        = string
  default     = "<open-exchange-rate-app-id>"
}
