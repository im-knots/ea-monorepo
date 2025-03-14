variable "dns_name" {
  type = string
}

variable "env" {
  type = string
}

variable "mgmt_project" {
  type = string
}

variable "nonprod_projects" {
  type = list(string)
}

