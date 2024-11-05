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

#################################################################################
# IAM for dcr-cluster-sa
#################################################################################
resource "google_project_iam_member" "dcr_cluster_sa_registry_reader" {
  project = var.project_id
  role    = "roles/artifactregistry.reader"
  member  = "serviceAccount:${google_service_account.gcp_dcr_cluster_sa.email}"
}

resource "google_project_iam_member" "dcr_cluster_sa_logger_writer" {
  project = var.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.gcp_dcr_cluster_sa.email}"
}

resource "google_project_iam_member" "dcr_cluster_sa_metric_writer" {
  project = var.project_id
  role    = "roles/monitoring.metricWriter"
  member  = "serviceAccount:${google_service_account.gcp_dcr_cluster_sa.email}"
}

resource "google_project_iam_member" "dcr_cluster_sa_metric_viewer" {
  project = var.project_id
  role    = "roles/monitoring.viewer"
  member  = "serviceAccount:${google_service_account.gcp_dcr_cluster_sa.email}"
}

#################################################################################
# IAM for dcr-cvm-sa
#################################################################################
resource "google_project_iam_member" "cvm_sa_log_writter" {
  project = var.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.gcp_cvm_sa.email}"
}

resource "google_project_iam_member" "cvm_sa_cc_workload_user" {
  project = var.project_id
  role    = "roles/confidentialcomputing.workloadUser"
  member  = "serviceAccount:${google_service_account.gcp_cvm_sa.email}"
}

resource "google_project_iam_member" "cvm_sa_storage_admin" {
  project = var.project_id
  role    = "roles/storage.objectAdmin"
  member  = "serviceAccount:${google_service_account.gcp_cvm_sa.email}"
}

resource "google_project_iam_member" "cvm_sa_registry_reader" {
  project = var.project_id
  role    = "roles/artifactregistry.reader"
  member  = "serviceAccount:${google_service_account.gcp_cvm_sa.email}"
}

#################################################################################
# IAM for dcr-pod-sa
#################################################################################
resource "google_project_iam_member" "dcr_pod_sa_storage_admin" {
  project = var.project_id
  role    = "roles/storage.objectAdmin"
  member  = "serviceAccount:${google_service_account.gcp_dcr_pod_sa.email}"
}

resource "google_project_iam_member" "dcr_pod_sa_sql_client" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.gcp_dcr_pod_sa.email}"
}

resource "google_project_iam_member" "dcr_pod_sa_log_writer" {
  project = var.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.gcp_dcr_pod_sa.email}"
}

resource "google_project_iam_member" "dcr_pod_sa_instance_admin" {
  project = var.project_id
  role    = "roles/compute.instanceAdmin.v1"
  member  = "serviceAccount:${google_service_account.gcp_dcr_pod_sa.email}"
}

resource "google_project_iam_member" "dcr_pod_sa_sauser" {
  project = var.project_id
  role    = "roles/iam.serviceAccountUser"
  member  = "serviceAccount:${google_service_account.gcp_dcr_pod_sa.email}"
}

resource "google_kms_key_ring_iam_member" "dcr_pod_sa_keyring_admin" {
  key_ring_id = google_kms_key_ring.dcr_key_ring.id
  role        = "roles/cloudkms.admin"
  member      = "serviceAccount:${google_service_account.gcp_dcr_pod_sa.email}"
}

resource "google_project_iam_member" "dcr_pod_sa_wip_admin" {
  project = var.project_id
  role    = "roles/iam.workloadIdentityPoolAdmin"
  member  = "serviceAccount:${google_service_account.gcp_dcr_pod_sa.email}"
}

resource "google_service_account_iam_member" "dcr_pod_sa_iam_member" {
  service_account_id = google_service_account.gcp_dcr_pod_sa.id
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.dcr-workload-identity-pool.name}/*"
}

resource "google_project_iam_member" "dcr_pod_sa_repo_writer" {
  project = var.project_id
  role    = "roles/artifactregistry.writer"
  member  = "serviceAccount:${google_service_account.gcp_dcr_pod_sa.email}"
}

#################################################################################
# IAM for jupyter-pod-sa
#################################################################################

resource "google_project_iam_member" "jupyter_pod_sa_logger_writer" {
  project = var.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.gcp_jupyter_pod_sa.email}"
}

resource "google_project_iam_member" "jupyter_pod_sa_metric_writer" {
  project = var.project_id
  role    = "roles/monitoring.metricWriter"
  member  = "serviceAccount:${google_service_account.gcp_jupyter_pod_sa.email}"
}

resource "google_project_iam_member" "jupyter_pod_sa_metric_viewer" {
  project = var.project_id
  role    = "roles/monitoring.viewer"
  member  = "serviceAccount:${google_service_account.gcp_jupyter_pod_sa.email}"
}