output "gke_service_account" {
  value = data.google_container_cluster.gke.node_config[0].service_account
}