terraform {
    required_version = "1.10.3"
    backend "gcs" {
        bucket  = "eru-labs-terraform-state-dev"
        prefix  = "eru-labs-infra/terraform.tfstate" # Optional: folder-like structure for state
    }
    required_providers {
        google = {
            source = "hashicorp/google"
            version = "6.14.1"
        }
    }
}

locals {
    project      = "eru-labs-dev-446616"
    region       = "us-central1"
    cluster_name = "eru-labs-dev"

    gcp_apis = toset([
        "artifactregistry.googleapis.com",
        "compute.googleapis.com",
    ])

    eru_services = toset([
    "eru-labs-brand-frontend", 
    "eru-labs-brand-backend", 
    "ea-platform-frontend", 
    "ea-platform-backend"
  ])

}

provider "google" {
  project     = local.project
  region      = local.region
}

module project_setup {
  source   = "../../modules/project"
  
  project  = local.project
  gcp_apis = local.gcp_apis

}


module artifact_registry {
  source  = "../../modules/artifact-registry"

  eru_services = local.eru_services

  depends_on = [
    module.project_setup
  ]

}

module gke {
  source = "../../modules/gke"
  
  cluster_name = local.cluster_name

  depends_on = [
    module.project_setup
  ]

}




