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

# A service account used for  data clean room cluster
resource "google_service_account" "gcp_dcr_cluster_sa" {
  account_id   = "dcr-${var.env}-cluster-sa"
  display_name = "A Service account for data clean room cluster"
  project      = var.project_id
}

resource "google_service_account" "gcp_cvm_sa" {
  account_id   = "dcr-${var.env}-cvm-sa"
  display_name = "A Service account for confidential vm"
  project      = var.project_id
}

resource "google_service_account" "gcp_dcr_pod_sa" {
  account_id   = "dcr-${var.env}-pod-sa"
  display_name = "A Service account for data clean room api pod"
  project      = var.project_id
}

resource "google_service_account" "gcp_jupyter_pod_sa" {
  account_id   = "jupyter-${var.env}-pod-sa"
  display_name = "A Service account for jupyterhub single user pod"
  project      = var.project_id
}