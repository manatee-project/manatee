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

resource "google_sql_database_instance" "dcr_database_instance" {
  name                = "dcr-${var.env}-db-instance"
  database_version    = "MYSQL_8_0"
  project             = var.project_id
  region              = var.region
  deletion_protection = false
  settings {
    tier = "db-f1-micro"

    ip_configuration {
      ipv4_enabled                                  = false
      private_network                               = google_compute_network.data_clean_room_network.id
      enable_private_path_for_google_cloud_services = true
    }
  }
  lifecycle {
    prevent_destroy = false
  }
  depends_on = [google_compute_subnetwork.data_clean_room_subnetwork, google_compute_global_address.dcr_private_address, google_service_networking_connection.private_vpc_connection]
}
