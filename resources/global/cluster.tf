/**
 * Copyright 2024 TikTok Pte. Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

locals {
  cluster_name   = "dcr-${var.env}-cluster"
  node_pool_name = "dcr-${var.env}-node-pool"
  node_pool_name_cpu = "dcr-${var.env}-node-pool-cpu"
}

# GKE Cluster
resource "google_container_cluster" "dcr_cluster" {
  project = var.project_id
  name    = local.cluster_name
  # if use region, each zone will create a node
  location = var.zone
  # We can't create a cluster with no node pool defined, but we want to only use
  # separately managed node pools. So we create the smallest possible default
  # node pool and immediately delete it.
  deletion_protection      = false
  remove_default_node_pool = true
  enable_l4_ilb_subsetting = true
  initial_node_count       = 1
  workload_identity_config {
    workload_pool = "${var.project_id}.svc.id.goog"
  }
  ip_allocation_policy {
    stack_type = "IPV4_IPV6"
  }
  datapath_provider = "ADVANCED_DATAPATH"
  network           = google_compute_network.data_clean_room_network.self_link
  subnetwork        = google_compute_subnetwork.data_clean_room_subnetwork.self_link
}


# Note pool for GKE cluster
resource "google_container_node_pool" "dcr_node_pool" {
  project    = var.project_id
  name       = local.node_pool_name
  location   = var.zone
  cluster    = google_container_cluster.dcr_cluster.name
  management {
    auto_repair = false
  }
  node_config {
    machine_type = var.machine_type
    confidential_nodes {
      enabled = true
      confidential_instance_type = "TDX"
    }
    # Add the GPU configuration.
    guest_accelerator {
      type  = var.gpu_type
      count = var.gpu_count
      gpu_driver_installation_config {
        gpu_driver_version = "INSTALLATION_DISABLED"
      }
    }
    reservation_affinity {
      consume_reservation_type = "NO_RESERVATION"
    }
    flex_start = true
    # GKE nodes with GPUs need this scope to install drivers.
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform",
    ]
    service_account = google_service_account.gcp_dcr_cluster_sa.email
    max_run_duration = "14400s"
  }

  depends_on = [
    google_service_account.gcp_dcr_cluster_sa,
  ]
  autoscaling {
    max_node_count = 3
    min_node_count = 0
  }
}

resource "google_container_node_pool" "dcr_node_pool2" {
  project    = var.project_id
  name       = local.node_pool_name_cpu
  location   = var.zone
  cluster    = google_container_cluster.dcr_cluster.name
  node_count = var.num_nodes

  node_config {
    service_account = google_service_account.gcp_dcr_cluster_sa.email
    preemptible     = false
    machine_type    = var.cpu_machine_type
  }

  depends_on = [
    google_service_account.gcp_dcr_cluster_sa,
  ]
  autoscaling {
    max_node_count = 3
  }
}
