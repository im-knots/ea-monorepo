variable "dns_name" {
  type = string
}

variable "env" {
  type = string
}

variable "project" {
  type = string
}

variable "delegated_users" {
  type = list(string)
}

variable "delegated_nameservers" {
  type = map(list(string))
  default = {}
}