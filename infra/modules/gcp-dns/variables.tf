variable "dns_name" {
  type = string
}

variable "env" {
  type = string
}

variable "delegated_nameservers" {
  type = map(list(string))
  default = {}
}