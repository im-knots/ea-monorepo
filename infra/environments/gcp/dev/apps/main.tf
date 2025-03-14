terraform {
    required_version = "1.10.3"
    backend "gcs" {
        bucket  = "eru-labs-terraform-state-dev"
        prefix  = "eru-labs-apps-dev/terraform.tfstate" # Optional: folder-like structure for state
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
    env          = "dev"

    namespaces   = toset(["ea-platform", "eru-labs-brand"])

    eru_apps = {
        "brand-backend" = {
            chart = "oci://us-central1-docker.pkg.dev/eru-labs-mgmt/eru-labs-images/brand-backend-chart"
            version = "20250312-d70497f"
            helm_overrides = {
                "image.repository"      = "us-central1-docker.pkg.dev/eru-labs-mgmt/eru-labs-images/brand-backend"
                "image.tag"             = "20250312-d70497f"
                "gke.enabled"           = "true"
                
                "ingress.className"                  = "gce"
                "ingress.hosts[0].host"              = "backend.dev.erulabs.ai"
                "ingress.hosts[0].paths[0].path"     = "/"
                "ingress.hosts[0].paths[0].pathType" = "Prefix"
                "ingress.annotations.kubernetes\\.io/ingress\\.class" = "gce"
                "ingress.annotations.networking\\.gke\\.io/managed-certificates" = "brand-frontend-cert"
            }
        }
        "brand-frontend" = {
            chart = "oci://us-central1-docker.pkg.dev/eru-labs-mgmt/eru-labs-images/brand-frontend-chart"
            version = "20250312-d70497f"
            helm_overrides = {
                "image.repository"      = "us-central1-docker.pkg.dev/eru-labs-mgmt/eru-labs-images/brand-frontend"
                "image.tag"             = "20250312-d70497f"
                "config.apiUrl"         = "https://backend.dev.erulabs.ai"
                "gke.enabled"           = "true"
                
                "ingress.className"                  = "gce"
                "ingress.hosts[0].host"              = "dev.erulabs.ai"
                "ingress.hosts[0].paths[0].path"     = "/"
                "ingress.hosts[0].paths[0].pathType" = "Prefix"
                "ingress.annotations.kubernetes\\.io/ingress\\.class" = "gce"
                "ingress.annotations.networking\\.gke\\.io/managed-certificates" = "brand-frontend-cert"
            }
        }
    }

    ea_apps = {
        "ea-agent-manager" = {
            chart = "oci://us-central1-docker.pkg.dev/eru-labs-mgmt/eru-labs-images/ea-agent-manager-chart"
            version = "20250312-d70497f"
            helm_overrides = {
                "image.repository"  = "us-central1-docker.pkg.dev/eru-labs-mgmt/eru-labs-images/ea-agent-manager"
                "image.tag"         = "20250312-d70497f"
            }
        }
        "ea-ainu-manager" = {
            chart = "oci://us-central1-docker.pkg.dev/eru-labs-mgmt/eru-labs-images/ea-ainu-manager-chart"
            version = "20250312-d70497f"
            helm_overrides = {
                "image.repository"  = "us-central1-docker.pkg.dev/eru-labs-mgmt/eru-labs-images/ea-ainu-manager"
                "image.tag"         = "20250312-d70497f"
            }
        }
        "ea-ainu-operator" = {
            chart = "oci://us-central1-docker.pkg.dev/eru-labs-mgmt/eru-labs-images/ea-ainu-operator-chart"
            version = "20250312-d70497f"
            helm_overrides = {
                "image.repository"  = "us-central1-docker.pkg.dev/eru-labs-mgmt/eru-labs-images/ea-ainu-operator"
                "image.tag"         = "20250312-d70497f"
            }
        }
        "ea-credentials-manager" = {
            chart = "oci://us-central1-docker.pkg.dev/eru-labs-mgmt/eru-labs-images/ea-credentials-manager-chart"
            version = "20250312-d70497f"
            helm_overrides = {
                "image.repository"  = "us-central1-docker.pkg.dev/eru-labs-mgmt/eru-labs-images/ea-credentials-manager"
                "image.tag"         = "20250312-d70497f"
            }
        }
        "ea-front" = {
            chart = "oci://us-central1-docker.pkg.dev/eru-labs-mgmt/eru-labs-images/ea-front-chart"
            version = "20250312-d70497f"
            helm_overrides = {
                "image.repository"  = "us-central1-docker.pkg.dev/eru-labs-mgmt/eru-labs-images/ea-front"
                "image.tag"         = "20250312-d70497f"
            }
        }
        "ea-job-api" = {
            chart = "oci://us-central1-docker.pkg.dev/eru-labs-mgmt/eru-labs-images/ea-job-api-chart"
            version = "20250312-d70497f"
            helm_overrides = {
                "image.repository"  = "us-central1-docker.pkg.dev/eru-labs-mgmt/eru-labs-images/ea-job-api"
                "image.tag"         = "20250312-d70497f"
            }
        }
        "ea-job-operator" = {
            chart = "oci://us-central1-docker.pkg.dev/eru-labs-mgmt/eru-labs-images/ea-job-operator-chart"
            version = "20250312-d70497f"
            helm_overrides = {
                "image.repository"  = "us-central1-docker.pkg.dev/eru-labs-mgmt/eru-labs-images/ea-job-operator"
                "image.tag"         = "20250312-d70497f"
            }
        }
        "ea-job-utils" = {
            chart = "oci://us-central1-docker.pkg.dev/eru-labs-mgmt/eru-labs-images/ea-job-utils-chart"
            version = "20250312-d70497f"
            helm_overrides = {
                "image.repository"  = "us-central1-docker.pkg.dev/eru-labs-mgmt/eru-labs-images/ea-job-utils"
                "image.tag"         = "20250312-d70497f"
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



