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

resource "google_sql_database" "database" {
  name     = "dcr-${var.namespace}-database"
  project  = var.project_id
  instance = "dcr-${var.env}-db-instance"
}

resource "google_sql_user" "dcr_db_user" {
  name     = var.mysql_username
  instance = "dcr-${var.env}-db-instance"
  password = var.mysql_password
  project  = var.project_id
}
