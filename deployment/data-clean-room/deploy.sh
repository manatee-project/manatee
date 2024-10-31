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

VAR_FILE="../../env.bzl"
if [ ! -f "$VAR_FILE" ]; then
    echo "Error: Variables file does not exist."
    exit 1
fi

VAR_FILE=$(realpath $VAR_FILE)
source $VAR_FILE

if [ -z "$1" ]
then 
    echo "Error: No namespace argument supplied."
    exit 1
fi
namespace=$1

tag="latest"
helm_name="data-clean-room-helm"

connection_name="${project_id}:${region}:dcr-${env}-db-instance"
service_account="dcr-k8s-pod-sa"
docker_repo="dcr-${env}-images"
api_docker_reference="us-docker.pkg.dev/${project_id}/${docker_repo}/data-clean-room-api"
monitor_docker_reference="us-docker.pkg.dev/${project_id}/${docker_repo}/data-clean-room-monitor"

helm upgrade --cleanup-on-fail \
    --set apiImage.repository=${api_docker_reference} \
    --set apiImage.tag=${tag} \
    --set monitorImage.repository=${monitor_docker_reference} \
    --set monitorImage.tag=${tag} \
    --set serviceAccount.name=${service_account} \
    --set cloudSql.connection_name=${connection_name} \
    --set namespace=${namespace} \
    --install $helm_name ./ \
    --namespace $namespace \
    --values config.yaml