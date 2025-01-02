variable eru_services {
  type        = set(string)
  default     = [""]
  description = "The list of services to create Docker Image Repos for"
}
