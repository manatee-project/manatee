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
  gcp_dcr_pod_sa_email     = "${local.gcp_dcr_pod_sa}@${var.project_id}.iam.gserviceaccount.com"
  gcp_jupyter_pod_sa_email = "${local.gcp_jupyter_pod_sa}@${var.project_id}.iam.gserviceaccount.com"
  admin_binding            = "${data.google_client_openid_userinfo.me.email}-binding"
}

resource "kubernetes_service_account" "k8s_dcr_pod_service_account" {
  metadata {
    name      = "dcr-k8s-pod-sa"
    namespace = var.namespace
    annotations = {
      "iam.gke.io/gcp-service-account" = local.gcp_dcr_pod_sa_email
    }
  }
  automount_service_account_token = true
  depends_on                      = [kubernetes_namespace.data_clean_room_k8s_namespace]
}

resource "kubernetes_service_account" "k8s_jupyter_pod_service_account" {
  metadata {
    name      = "jupyter-k8s-pod-sa"
    namespace = var.namespace
    annotations = {
      "iam.gke.io/gcp-service-account" = local.gcp_jupyter_pod_sa_email
    }
  }
  automount_service_account_token = true
  depends_on                      = [kubernetes_namespace.jupyterhub_k8s_namespace]
}


resource "google_service_account_iam_member" "dcr_pod_sa_iam_member" {
  service_account_id = "projects/${var.project_id}/serviceAccounts/${local.gcp_dcr_pod_sa}@${var.project_id}.iam.gserviceaccount.com"
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:${var.project_id}.svc.id.goog[${var.namespace}/${kubernetes_service_account.k8s_dcr_pod_service_account.metadata[0].name}]"
  depends_on         = [kubernetes_namespace.data_clean_room_k8s_namespace]
}

resource "google_service_account_iam_member" "jupyter_pod_sa_iam_member" {
  service_account_id = "projects/${var.project_id}/serviceAccounts/${local.gcp_jupyter_pod_sa}@${var.project_id}.iam.gserviceaccount.com"
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:${var.project_id}.svc.id.goog[${var.namespace}/${kubernetes_service_account.k8s_jupyter_pod_service_account.metadata[0].name}]"
  depends_on         = [kubernetes_namespace.data_clean_room_k8s_namespace]
}

resource "kubernetes_cluster_role_binding" "cluster_admin_binding" {
  metadata {
    name = local.admin_binding
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = "cluster-admin"
  }
  subject {
    kind      = "User"
    name      = data.google_client_openid_userinfo.me.email
    api_group = "rbac.authorization.k8s.io"
  }
}
