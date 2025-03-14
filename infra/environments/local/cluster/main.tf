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
  gateway_host = "*.erulabs.local"
  
  enable_sec_tooling = true
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

module "minikube" {
  source = "../../../modules/minikube-setup"
}

module monitoring {
  source = "../../../modules/monitoring"
  
  env                = local.env
  enable_sec_tooling = local.enable_sec_tooling
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

module "ea_gateway" {
  source       = "../../../modules/gateway"
  namespace    = "ea-platform"
  gateway_host = local.gateway_host
  depends_on   = [ module.k8s_namespace, module.istio ]
}

