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

  ea_apps = {
    "ea-agent-manager"       = "../../../../ea-platform/ea-agent-manager/chart"
    "ea-credentials-manager" = "../../../../ea-platform/ea-credentials-manager/chart"
    "ea-ainu-manager"        = "../../../../ea-platform/ea-ainu-manager/chart"
    "ea-ainu-operator"       = "../../../../ea-platform/ea-ainu-operator/chart"
    "ea-front"               = "../../../../ea-platform/ea-front/chart"
    "ea-job-api"             = "../../../../ea-platform/ea-job-api/chart"
    "ea-job-operator"        = "../../../../ea-platform/ea-job-operator/chart"
    "ea-job-utils"           = "../../../../ea-platform/ea-job-utils/chart"
  }

  eru_apps = {
    "brand-backend"  = "../../../../brand/brand-backend/chart"
    "brand-frontend" = "../../../../brand/brand-frontend/chart"
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

module "ea_app_deployment" {
  for_each = local.ea_apps

  source     = "../../../modules/app-deployment"
  app_name   = each.key
  chart_path = each.value
  namespace  = "ea-platform"

  depends_on = [ 
    module.mongodb_deployment,
  ]
}

module ollama {
  source    = "../../../modules/ollama"
  namespace = "ea-platform"
}

module "eru_labs_brand_app_deployment" {
  for_each = local.eru_apps

  source     = "../../../modules/app-deployment"
  app_name   = each.key
  chart_path = each.value
  namespace  = "eru-labs-brand"

  depends_on = [ 
    module.mongodb_deployment,
  ]
}