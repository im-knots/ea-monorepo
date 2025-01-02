variable project {
  type        = string
  default     = ""
  description = "The GCP project to target"
}


variable gcp_apis {
  type        = set(string)
  default     = [""]
  description = "the list of GCP APIs to enable for the project"
}
