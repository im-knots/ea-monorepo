# resource "google_artifact_registry_repository_iam_binding" "binding" {
#   for_each   = toset(["eru-labs-brand-frontend", "eru-labs-brand-backend", "ea-platform-frontend", "ea-platform-backend"])
#   project    = google_artifact_registry_repository.my-repo.project
#   location   = google_artifact_registry_repository.my-repo.location
#   repository = google_artifact_registry_repository.my-repo.name
#   role       = "roles/artifactregistry.reader"
#   members    = [
#     "user:jane@example.com",
#   ]
# }