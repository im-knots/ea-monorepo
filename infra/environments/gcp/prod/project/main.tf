terraform {
    required_version = "1.10.3"
    backend "gcs" {
        bucket  = "eru-labs-terraform-state-prod"
        prefix  = "eru-labs-project-prod/terraform.tfstate" # Optional: folder-like structure for state
    }
    required_providers {
        google = {
            source  = "hashicorp/google"
            version = "6.14.1"
        }
        helm = {
            source  = "hashicorp/helm"
            version = "2.17.0"
        }
        kubernetes = {
            source  = "hashicorp/kubernetes"
            version = "2.35.1"
        }
    }
}

locals {
    project      = "eru-labs-prod-453516"
    region       = "us-central1"
    env          = "prod"

    gcp_apis = toset([
        "compute.googleapis.com",
        "container.googleapis.com",
        "containerscanning.googleapis.com",
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

