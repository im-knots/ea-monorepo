resource "google_artifact_registry_repository" "source_artifact_repo" {
  for_each      = var.eru_services
  location      = "us-central1"
  repository_id = each.value
  description   = "docker image repository for ${each.value}"
  format        = "DOCKER"

  docker_config {
    immutable_tags = true
  }
}

resource "google_artifact_registry_repository" "helm_artifact_repo" {
  for_each      = var.eru_services
  location      = "us-central1"
  repository_id = "${each.value}-helm"
  description   = "docker image repository for ${each.value}"
  format        = "DOCKER"

  docker_config {
    immutable_tags = true
  }
}