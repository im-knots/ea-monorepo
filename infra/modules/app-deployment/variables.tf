# variable "app_name" {
#   description = "Name of the application to deploy"
#   type        = string
# }

# variable "chart_path" {
#   description = "Path to the local Helm chart for the application"
#   type        = string
# }

variable "namespace" {
  description = "The namespace to deploy the applications to"
  type        = string
}

# variable "helm_overrides" {
#   type = map(string)
# }

variable "apps" {
  description = "A map of applications with their Helm configurations"
  type = map(object({
    chart          = string
    version        = string
    helm_overrides = map(string)
  }))
}
