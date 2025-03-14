resource "google_dns_managed_zone" "zone" {
  name     = "erulabs-public-zone"
  dns_name = "${var.dns_name}." # Change this to your domain
  description = "Public DNS zone for ${var.dns_name}"

  visibility = "public"

}

resource "google_dns_managed_zone" "zones_nonprod" {
  for_each = toset(["dev", "stage"]) 
  name     = "erulabs-public-zone-${each.key}"
  dns_name = "${each.key}.${var.dns_name}." # Change this to your domain
  description = "Public DNS zone for ${each.key}.${var.dns_name}"

  visibility = "public"
}

resource "google_dns_record_set" "dev_ns" {
  for_each     = toset(["dev", "stage"]) 
  name         = "${each.key}.${var.dns_name}."
  type         = "NS"
  ttl          = 300
  managed_zone = google_dns_managed_zone.zone.name

  rrdatas = google_dns_managed_zone.zones_nonprod[each.key].name_servers
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

resource "google_service_account" "dns_admin_account" {
  account_id   = "dns-update"
  display_name = "dns update service account for external-dns"
}

resource "google_project_iam_binding" "dns_admin_account" {
  project = var.mgmt_project
  role    = "roles/dns.admin"

  members = [
    "serviceAccount:${google_service_account.dns_admin_account.email}",
  ]
}

resource "google_service_account" "dns_wi_account" {
  for_each     = toset(var.nonprod_projects)
  project      = each.key
  account_id   = "external-dns"
  display_name = "Workload Identity Service Account for ExternalDNS"
}

resource "google_project_iam_binding" "dns_wi_account_read" {
  for_each = toset(var.nonprod_projects)
  project  = var.mgmt_project
  role     = "roles/dns.reader"

  members = [
    "serviceAccount:${google_service_account.dns_wi_account[each.key].email}",
  ]
}

resource "google_service_account_iam_binding" "workload_identity_binding" {
  for_each           = toset(var.nonprod_projects)
  service_account_id = google_service_account.dns_wi_account[each.key].name
  role               = "roles/iam.workloadIdentityUser"

  members = [
    "serviceAccount:${each.key}.svc.id.goog[external-dns/external-dns]",
  ]
}

resource "google_service_account_iam_binding" "allow_dns_updates" {
  for_each           = toset(var.nonprod_projects)
  service_account_id = google_service_account.dns_admin_account.name
  role               = "roles/iam.serviceAccountTokenCreator"

  members = [
    "serviceAccount:${google_service_account.dns_wi_account[each.key].email}",
  ]
}