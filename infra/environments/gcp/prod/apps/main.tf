terraform {
    required_version = "1.10.3"
    backend "gcs" {
        bucket  = "eru-labs-terraform-state-prod"
        prefix  = "eru-labs-apps-prod/terraform.tfstate" # Optional: folder-like structure for state
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
    cluster_name = "eru-labs-prod"
    env          = "prod"

    namespaces   = toset(["ea-platform", "eru-labs-brand"])

    eru_apps = {
        "brand-backend" = {
            chart = "oci://us-central1-docker.pkg.dev/eru-labs-mgmt/eru-labs-images/brand-backend-chart"
            version = "20250316-befc6d9"
            helm_overrides = {
                "image.repository"      = "us-central1-docker.pkg.dev/eru-labs-mgmt/eru-labs-images/brand-backend"
                "image.tag"             = "20250316-befc6d9"
                "gke.enabled"           = "true"
                
                "ingress.className"                  = "gce"
                "ingress.hosts[0].host"              = "backend.erulabs.ai"
                "ingress.hosts[0].paths[0].path"     = "/"
                "ingress.hosts[0].paths[0].pathType" = "Prefix"
                "ingress.annotations.kubernetes\\.io/ingress\\.class" = "gce"
                "ingress.annotations.networking\\.gke\\.io/managed-certificates" = "brand-backend-cert"
                "ingress.annotations.networking\\.gke\\.io/v1beta1\\.FrontendConfig" = "brand-backend-redirect-http"
            }
        }
        "brand-frontend" = {
            chart = "oci://us-central1-docker.pkg.dev/eru-labs-mgmt/eru-labs-images/brand-frontend-chart"
            version = "20250316-befc6d9"
            helm_overrides = {
                "image.repository"      = "us-central1-docker.pkg.dev/eru-labs-mgmt/eru-labs-images/brand-frontend"
                "image.tag"             = "20250316-befc6d9"
                "config.apiUrl"         = "https://backend.erulabs.ai"
                "gke.enabled"           = "true"
                
                "ingress.className"                  = "gce"
                "ingress.hosts[0].host"              = "erulabs.ai"
                "ingress.hosts[0].paths[0].path"     = "/"
                "ingress.hosts[0].paths[0].pathType" = "Prefix"
                "ingress.annotations.kubernetes\\.io/ingress\\.class" = "gce"
                "ingress.annotations.networking\\.gke\\.io/managed-certificates" = "brand-frontend-cert"
                "ingress.annotations.networking\\.gke\\.io/v1beta1\\.FrontendConfig" = "brand-frontend-redirect-http"
            }
        }
    }

}

provider "google" {
  project     = local.project
  region      = local.region
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

module "mongodb_deployment" {
  for_each   = local.namespaces
  source     = "../../../../modules/mongodb"
  namespace  = each.key
}

module "eru_app_deployment" {
  source     = "../../../../modules/app-deployment"
  apps       = local.eru_apps
  namespace  = "eru-labs-brand"

  depends_on = [ 
    module.mongodb_deployment,
  ]
}

# module "ea_app_deployment" {
#   source     = "../../../../modules/app-deployment"
#   apps       = local.ea_apps
#   namespace  = "ea-platform"

#   depends_on = [ 
#     module.mongodb_deployment,
#   ]
# }



