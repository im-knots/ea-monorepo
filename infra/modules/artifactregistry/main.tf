# Artifact Registry Docker repository
resource "google_artifact_registry_repository" "docker_repo" {
  location      = var.region
  repository_id = "eru-labs-images" # customize as needed
  description   = "Docker images for Eru Labs"
  format        = "DOCKER"
}

# IAM Binding to allow the service account to push images
resource "google_artifact_registry_repository_iam_member" "docker_push" {
  location   = google_artifact_registry_repository.docker_repo.location
  repository = google_artifact_registry_repository.docker_repo.name
  role       = "roles/artifactregistry.writer"
  member     = "serviceAccount:${var.service_account_email}"
}
