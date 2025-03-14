output "prod_nameservers" {
  value = google_dns_managed_zone.zone.name_servers
}
