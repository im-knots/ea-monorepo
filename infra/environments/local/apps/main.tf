terraform {
  required_version = ">= 1.10.3"

  backend "local" {}

  required_providers {
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
  cluster_name = "minikube"
  env          = "local"
  namespaces   = toset(["ea-platform", "eru-labs-brand"])

    eru_apps = {
        "brand-backend" = {
            chart = "../../../../brand/brand-backend/chart"
            version = null
            helm_overrides = {}
        }
        "brand-frontend" = {
            chart = "../../../../brand/brand-frontend/chart"
            version = null
            helm_overrides = {}
        }
    }

    ea_apps = {
        "ea-agent-manager" = {
            chart = "../../../../ea-platform/ea-agent-manager/chart"
            version = null
            helm_overrides = {}
        }
        "ea-ainu-manager" = {
            chart = "../../../../ea-platform/ea-ainu-manager/chart"
            version = null
            helm_overrides = {}
        }
        "ea-ainu-operator" = {
            chart = "../../../../ea-platform/ea-ainu-operator/chart"
            version = null
            helm_overrides = {}
        }
        "ea-credentials-manager" = {
            chart = "../../../../ea-platform/ea-credentials-manager/chart"
            version = null
            helm_overrides = {}
        }
        "ea-front" = {
            chart = "../../../../ea-platform/ea-front/chart"
            version = null
            helm_overrides = {
              "secrets.MONGO_URI" = "mongodb://mongodb.ea-platform:27017/ea"
              "jwks.fromSecret"   = "ea-front-jwks"
            }
        }
        "ea-job-api" = {
            chart = "../../../../ea-platform/ea-job-api/chart"
            version = null
            helm_overrides = {}
        }
        "ea-job-operator" = {
            chart = "../../../../ea-platform/ea-job-operator/chart"
            version = null
            helm_overrides = {}
        }
        "ea-job-utils" = {
            chart = "../../../../ea-platform/ea-job-utils/chart"
            version = null
            helm_overrides = {}
        }
    }
}

provider "kubernetes" {
  config_path    = "~/.kube/config"
  config_context = local.cluster_name
}

provider "helm" {
  kubernetes {
    config_path    = "~/.kube/config"
    config_context = local.cluster_name
  }
}

module "mongodb_deployment" {
  for_each   = local.namespaces
  source     = "../../../modules/mongodb"
  namespace  = each.key
}

module "eru_app_deployment" {
  source     = "../../../modules/app-deployment"
  apps       = local.eru_apps
  namespace  = "eru-labs-brand"

  depends_on = [ 
    module.mongodb_deployment,
  ]
}

# module ollama {
#   source    = "../../../modules/ollama"
#   namespace = "ea-platform"
# }

module "eru_labs_brand_app_deployment" {
  source     = "../../../modules/app-deployment"
  apps       = local.ea_apps
  namespace  = "ea-platform"

  depends_on = [ 
    module.mongodb_deployment,
  ]
}

resource "kubernetes_manifest" "ea-front-jwks" {
  manifest = {
    apiVersion = "v1"
    kind        = "Secret"
    metadata = {
      name      = "ea-front-jwks"
      namespace = "ea-platform"
    }
    data = {
      "jwks.json" = filebase64("./jwks/jwks.json")
      "private.json" = filebase64("./jwks/private.json")
    }
  }
}