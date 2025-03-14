terraform {
    required_version = "1.10.3"
    backend "gcs" {
        bucket  = "eru-labs-terraform-state-mgmt"
        prefix  = "eru-labs-project-mgmt/terraform.tfstate" # Optional: folder-like structure for state
    }
    required_providers {
        google = {
            source  = "hashicorp/google"
            version = "6.14.1"
        }
    }
}

locals {
    project      = "eru-labs-mgmt"
    env          = "mgmt"
    region       = "us-central1"
    dns_name     = "erulabs.ai"

    nonprod_projects = ["eru-labs-dev-446616"]

    gcp_apis = toset([
        "artifactregistry.googleapis.com",
        "iam.googleapis.com",
        "sts.googleapis.com",
        "dns.googleapis.com",
    ])
    artifact_pull_service_accounts = [
      "135550390903-compute@developer.gserviceaccount.com", #Dev Project
    ]
}

provider "google" {
  project     = local.project
  region      = local.region
}

module project_apis {
  source   = "../../../../modules/project-apis"
  
  project  = local.project
  gcp_apis = local.gcp_apis

}

module "gh_actions_workload_idenity" {
  source = "../../../../modules/gh-workload-idenity"

  project = local.project
  depends_on = [ module.project_apis ]
}

module "artifactregistry" {
  source = "../../../../modules/artifactregistry"

  region = local.region
  service_account_email = module.gh_actions_workload_idenity.service_account_email
  artifact_pull_service_accounts = local.artifact_pull_service_accounts
}

module "dns_zone" {
  source = "../../../../modules/gcp-dns"

  mgmt_project          = local.project 
  nonprod_projects      = local.nonprod_projects
  dns_name              = local.dns_name
  env                   = local.env
  depends_on            = [ module.project_apis ]
}