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

    gcp_apis = toset([
        "artifactregistry.googleapis.com",
        "iam.googleapis.com",
        "sts.googleapis.com",
        "dns.googleapis.com",
    ])

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
}

module "dns_zone" {
  source = "../../../../modules/gcp-dns"

  dns_name    = local.dns_name
  env         = local.env
  depends_on  = [ module.project_apis ]
}