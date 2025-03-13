resource "google_dns_managed_zone" "zone" {
  name     = "erulabs-public-zone"
  dns_name = "${var.dns_name}." # Change this to your domain
  description = "Public DNS zone for ${var.dns_name}"

  visibility = "public"

}

resource "google_dns_record_set" "mx_record" {
  count = var.env == "mgmt" ? 1 : 0
  name         = "erulabs.ai."
  type         = "MX"
  ttl          = 300
  managed_zone = google_dns_managed_zone.zone.name

  rrdatas = [
    "1 SMTP.GOOGLE.COM."
  ]

  lifecycle {
    prevent_destroy = true # This is our google workspaces email MX record so we cant destroy it or we will break email
  }
}