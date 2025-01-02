terraform {
    required_version = "1.10.3"
    backend "gcs" {
        bucket  = "eru-labs-terraform-state-dev"
        prefix  = "eru-labs-infra/terraform.tfstate" # Optional: folder-like structure for state
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
    project      = "eru-labs-dev-446616"
    region       = "us-central1"
    cluster_name = "eru-labs-dev"

    gcp_apis = toset([
        "artifactregistry.googleapis.com",
        "compute.googleapis.com",
        "container.googleapis.com",
        "containerscanning.googleapis.com",
    ])

    eru_services = toset([
    "eru-labs-brand-frontend", 
    "eru-labs-brand-backend", 
    "ea-platform-frontend", 
    "ea-platform-backend",
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

data "google_container_cluster" "cluster" {
    name     = local.cluster_name
    location = local.region
}

data "google_client_config" "provider" {}

provider "kubernetes" {
    host  = "https://${data.google_container_cluster.cluster.endpoint}"
    token = data.google_client_config.provider.access_token
    cluster_ca_certificate = base64decode(data.google_container_cluster.cluster.master_auth[0].cluster_ca_certificate) 
}

provider "helm" {
    kubernetes {
        host  = "https://${data.google_container_cluster.cluster.endpoint}"
        token = data.google_client_config.provider.access_token
        cluster_ca_certificate = base64decode(data.google_container_cluster.cluster.master_auth[0].cluster_ca_certificate) 
    }
}

module monitoring {
  source = "../../modules/monitoring"
  
  depends_on = [
    module.gke
  ]
}
