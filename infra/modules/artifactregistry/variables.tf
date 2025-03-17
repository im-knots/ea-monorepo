variable "region" {
  type = string
  default = "us-central1"
}

variable "service_account_email" {
  type = string
}

variable "artifact_pull_service_accounts" {
  type = list(string)
}