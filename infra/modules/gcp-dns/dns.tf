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

resource "google_project_iam_binding" "external_dns" {
  for_each = var.env == "dev" ? toset(var.delegated_users) : toset([])
  project  = var.project
  role     = "roles/dns.admin"

  members = [
    "serviceAccount:${each.key}",
  ]
}

data "google_iam_policy" "admin" {
  for_each = var.env == "dev" ? toset(var.delegated_users) : toset([])
  binding {
    role = "roles/viewer"
    members = [
      "serviceAccount:${each.key}",
    ]
  }
}

resource "google_dns_managed_zone_iam_policy" "policy" {
  for_each     = var.env == "dev" ? toset(var.delegated_users) : toset([])
  project      = var.project
  managed_zone = google_dns_managed_zone.zone.name
  policy_data  = data.google_iam_policy.admin[each.key].policy_data
}

resource "google_dns_record_set" "dev_ns" {
  for_each     = var.env == "mgmt" ? var.delegated_nameservers : {}
  name         = each.key
  type         = "NS"
  ttl          = 300
  managed_zone = google_dns_managed_zone.zone.name

  rrdatas = each.value
}