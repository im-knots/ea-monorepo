# Identity Pool
resource "google_iam_workload_identity_pool" "github_pool" {
  project                   = var.project
  workload_identity_pool_id = "actions-pool"
  display_name              = "GitHub Actions Pool"
}

# Provider within the pool for GitHub Actions (OIDC)
resource "google_iam_workload_identity_pool_provider" "github" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.github_pool.workload_identity_pool_id
  workload_identity_pool_provider_id = "github-provider"
  display_name                       = "Github Provider"
  description                        = "GitHub Actions identity pool provider for image build + push"
  disabled                           = false
  attribute_condition = <<EOT
    attribute.repository == "eru-labs/eru-labs-monorepo"
EOT
  attribute_mapping = {
    "google.subject"       = "assertion.sub"
    "attribute.actor"      = "assertion.actor"
    "attribute.aud"        = "assertion.aud"
    "attribute.repository" = "assertion.repository"
  }
  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }
}

resource "google_service_account" "gh_actions" {
  account_id   = "github-actions"
  display_name = "GitHub Actions Service Account"
}

resource "google_service_account_iam_member" "github_actions_impersonation" {
  service_account_id = google_service_account.gh_actions.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github_pool.name}/attribute.repository/eru-labs/eru-labs-monorepo"
}

resource "google_service_account_iam_member" "github_actions_service_account_user" {
  service_account_id = google_service_account.gh_actions.name
  role               = "roles/iam.serviceAccountUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github_pool.name}/attribute.repository/eru-labs/eru-labs-monorepo"
}

resource "google_service_account_iam_member" "github_actions_token_creator" {
  service_account_id = google_service_account.gh_actions.name
  role               = "roles/iam.serviceAccountTokenCreator"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github_pool.name}/attribute.repository/eru-labs/eru-labs-monorepo"
}
