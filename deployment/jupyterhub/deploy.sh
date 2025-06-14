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
helm_name="jupyterhub-helm-$namespace"
api="http://manatee.$namespace.svc.cluster.local"

service_account="jupyter-k8s-pod-sa"
docker_repo="dcr-${env}-${namespace}-images"
docker_reference="us-docker.pkg.dev/${project_id}/${docker_repo}/manatee-jupyterlab-singleuser"

helm repo add jupyterhub https://hub.jupyter.org/helm-chart/
helm repo update

helm upgrade --cleanup-on-fail \
    --set singleuser.image.name=${docker_reference} \
    --set singleuser.image.tag=${tag} \
    --set singleuser.serviceAccountName=${service_account} \
    --set singleuser.extraEnv.DATA_CLEAN_ROOM_HOST=${api} \
    --set singleuser.extraEnv.EXECUTION_STAGE='"1"' \
    --set singleuser.extraEnv.MANATEE_EXTRA_ENV_EXECUTION_STAGE='"2"' \
    --set singleuser.extraEnv.DEPLOYMENT_ENV=${env} \
    --set singleuser.extraEnv.PROJECT_ID=${project_id} \
    --set singleuser.extraEnv.KEY_LOCALTION=${region} \
    --set singleuser.networkPolicy.enabled=false \
    --set singleuser.storage.capacity=20Gi \
    --install $helm_name jupyterhub/jupyterhub \
    --namespace ${namespace} \
    --version=3.0.3 \
    --values config.yaml

echo "Deployment Completed."
echo "Try 'kubectl --namespace=$namespace get service proxy-public' to obtain external IP"
