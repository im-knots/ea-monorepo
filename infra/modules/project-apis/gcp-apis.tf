resource "google_project_service" "apis_enable" {
  for_each = var.gcp_apis
  project  = var.project
  service  = each.value

  disable_on_destroy = false
}