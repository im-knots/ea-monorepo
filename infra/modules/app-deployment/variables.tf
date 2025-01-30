variable "app_name" {
  description = "Name of the application to deploy"
  type        = string
}

variable "chart_path" {
  description = "Path to the local Helm chart for the application"
  type        = string
}

variable "namespace" {
  description = "The namespace to deploy the applications to"
  type        = string
}