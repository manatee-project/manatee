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

resource "google_compute_network" "data_clean_room_network" {
  name                    = "dcr-${var.env}-network"
  auto_create_subnetworks = false
  project                 = var.project_id
}

resource "google_compute_global_address" "dcr_private_address" {
  name          = "dcr-${var.env}-private-address"
  project       = var.project_id
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.data_clean_room_network.self_link
}

resource "google_service_networking_connection" "private_vpc_connection" {
  network                 = google_compute_network.data_clean_room_network.self_link
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.dcr_private_address.name]
  depends_on = [ google_compute_global_address.dcr_private_address ]
}

resource "google_compute_subnetwork" "data_clean_room_subnetwork" {
  name          = "dcr-${var.env}-subnetwork"
  project       = var.project_id
  ip_cidr_range = "10.0.0.0/22"
  region        = var.region

  stack_type       = "IPV4_IPV6"
  ipv6_access_type = "EXTERNAL"

  network = google_compute_network.data_clean_room_network.id
  depends_on = [ google_service_networking_connection.private_vpc_connection ]
}
