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

resource "kubernetes_service_account" "k8s_dcr_pod_service_account" {
  metadata {
    name      = "dcr-k8s-pod-sa"
    namespace = var.namespace
  }
  automount_service_account_token = true
  depends_on                      = [kubernetes_namespace.data_clean_room_k8s_namespace]
}

resource "kubernetes_service_account" "k8s_jupyter_pod_service_account" {
  metadata {
    name      = "jupyter-k8s-pod-sa"
    namespace = var.namespace
  }
  automount_service_account_token = true
  depends_on                      = [kubernetes_namespace.data_clean_room_k8s_namespace]
}
