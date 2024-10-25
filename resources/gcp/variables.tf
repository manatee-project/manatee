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

variable "username" {
  type        = string
  description = "Username to postfix resources with"
  default     = ""
}

variable "env" {
  type        = string
  description = "Deployment environment, e.g., dev, prod, oss"
}

variable "region" {
  type        = string
  description = "Region to create the gcp resources"
}

variable "zone" {
  type        = string
  description = "Zone to create the gcp resources"
}

variable "project_id" {
  type        = string
  description = "The GCP project ID"
}

variable "project_number" {
  type        = string
  description = "The GCP project number"
}

variable "mysql_username" {
  type        = string
  description = "Mysql username"
}

variable "mysql_password" {
  type        = string
  description = "Mysql password"
}

variable "type" {
  type        = string
  description = "Instance type for the GKE instances"
  default     = "n2-standard-4"
}

variable "num_nodes" {
  type        = number
  description = "Number of nodes to create in the GKE cluster"
  default     = 2
}
