#!/bin/bash
# Copyright 2024 TikTok Pte. Ltd.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Check if gcloud is installed
if ! [ -x "$(command -v gcloud)" ]; then
	echo "Error: gcloud is not installed." >&2
	exit 1
fi

# Check if gcloud logged in
if ! gcloud auth list | grep -q 'ACTIVE'; then
	echo "Error: No active gcloud account found." >&2
	exit 1
fi

# check whether variables has been set
VAR_FILE="../env.bzl"
if [ ! -f "$VAR_FILE" ]; then
    echo "Error: Variables file does not exist."
    exit 1
fi
VAR_FILE=$(realpath $VAR_FILE)
source $VAR_FILE

if ! gsutil ls gs://dcr-tf-state-$env > /dev/null 2>&1; then
  gsutil mb -l us gs://dcr-tf-state-$env
fi

pushd gcp

cat << EOF > backend.tf
terraform {
  backend "gcs" {
    bucket = "dcr-tf-state-$env"
    prefix = "cloud"
  }
}
EOF

cp $VAR_FILE terraform.tfvars
terraform init -reconfigure
terraform apply
popd 

zone=$region-a
# get kubernete cluster credentials
gcloud container clusters get-credentials dcr-$env-cluster --zone $zone --project $project_id

pushd kubernetes
cp $VAR_FILE terraform.tfvars
terraform init -reconfigure
terraform apply
popd 