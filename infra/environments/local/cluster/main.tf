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

module monitoring {
  source = "../../../modules/monitoring"
  env    = local.env
}

module "k8s_namespace" {
  for_each = local.namespaces
  source     = "../../../modules/k8s-namespace"
  namespace  = each.key
}

module "istio" {
  source = "../../../modules/istio"
  depends_on = [ module.k8s_namespace ]
}

resource helm_release "ea-gateway" {
  name      = "ea-gateway"
  chart     = "../../../charts/gateway"
  namespace = "ea-platform"
  depends_on = [ module.k8s_namespace, module.istio ]
}