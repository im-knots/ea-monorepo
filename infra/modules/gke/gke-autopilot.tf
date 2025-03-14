resource "google_compute_network" "eru_labs" {
  name = "${var.cluster_name}-network"

  auto_create_subnetworks  = false
  enable_ula_internal_ipv6 = true
}

resource "google_compute_subnetwork" "eru_labs" {
  name = "${var.cluster_name}-subnetwork"

  ip_cidr_range = "10.0.0.0/16"
  region        = "us-central1"

  stack_type       = "IPV4_IPV6"
  ipv6_access_type = "INTERNAL"

  network = google_compute_network.eru_labs.id

  secondary_ip_range {
    range_name    = "services-range"
    ip_cidr_range = "192.168.0.0/24"
  }

  secondary_ip_range {
    range_name    = "pod-ranges"
    ip_cidr_range = "192.168.16.0/20" 
  }
}

resource "google_container_cluster" "gke" {
  name = var.cluster_name

  location                 = "us-central1"
  enable_autopilot         = true
  enable_l4_ilb_subsetting = true
  allow_net_admin          = true

  network    = google_compute_network.eru_labs.id
  subnetwork = google_compute_subnetwork.eru_labs.id

  ip_allocation_policy {
    stack_type                    = "IPV4_IPV6"
    services_secondary_range_name = google_compute_subnetwork.eru_labs.secondary_ip_range[0].range_name
    cluster_secondary_range_name  = google_compute_subnetwork.eru_labs.secondary_ip_range[1].range_name
  }

  # Set `deletion_protection` to `true` will ensure that one cannot
  # accidentally delete this instance by use of Terraform.
  deletion_protection = false
}