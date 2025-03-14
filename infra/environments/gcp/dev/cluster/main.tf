terraform {
    required_version = "1.10.3"
    backend "gcs" {
        bucket  = "eru-labs-terraform-state-dev"
        prefix  = "eru-labs-cluster-dev/terraform.tfstate" # Optional: folder-like structure for state
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
    
    mgmt_project        = "eru-labs-mgmt"
    public_domain       = "dev.erulabs.ai"
    public_zone_id      = "erulabs-public-zone-dev"
    dns_service_account = "external-dns@eru-labs-dev-446616.iam.gserviceaccount.com"


    enable_sec_tooling = false
}

provider "google" {
  project     = local.project
  region      = local.region
}

module gke {
  source = "../../../../modules/gke"
  
  cluster_name = local.cluster_name
}

data "google_container_cluster" "cluster" {
    name     = local.cluster_name
    location = local.region
    depends_on = [ module.gke ]
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

module "k8s_namespace" {
  for_each = local.namespaces
  source     = "../../../../modules/k8s-namespace"
  namespace  = each.key
}

module "external_dns" {
  source = "../../../../modules/external-dns"
  mgmt_project        = local.mgmt_project
  public_domain       = local.public_domain
  public_zone_id      = local.public_zone_id
  dns_service_account = local.dns_service_account
  cluster_name        = local.cluster_name
}

# module monitoring {
#   source = "../../../../modules/monitoring"
  
#   env                = local.env 
#   enable_sec_tooling = local.enable_sec_tooling
  
#   depends_on = [
#     module.gke
#   ]
# }
