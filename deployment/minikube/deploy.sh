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

env="minikube"
namespace="manatee"
project_id="mock-gcp-project-id"
region="us-west2"
zone="us-west2-a"
helm_name="manatee-helm"
eval $(minikube docker-env)
kubectl apply -f mysql-deployment.yaml -n $namespace
kubectl apply -f mysql-service.yaml -n $namespace
kubectl apply -f minio-dev.yaml
# deploy dcr api
helm upgrade --cleanup-on-fail \
    --set apiImage.repository=docker.io/library/api \
    --set apiImage.tag=latest \
    --set apiImage.pullPolicy=Never \
    --set monitorImage.repository=docker.io/library/reconciler \
    --set monitorImage.tag=latest \
    --set monitorImage.pullPolicy=Never \
    --set serviceAccount.name=dcr-k8s-pod-sa \
    --set serviceAccount.create=false \
    --set cloudSql.connection_name="" \
    --set namespace=${namespace} \
    --set config.env=${env} \
    --set config.projectId=${project_id} \
    --set config.zone=${zone} \
    --set config.region=${region} \
    --set config.debug=true \
    --set config.teeBackend=MOCK \
    --set config.registryType=MINIKUBE \
    --set config.storageType=MINIO \
    --set config.minioSecretKey=minioadmin \
    --set config.minioAccessKey=minioadmin \
    --set config.minioEndpoint=minio-service:9000 \
    --set mysql.host=mysql-service \
    --set mysql.port=3306 \
    --set useMinikube=true \
    --install $helm_name ../manatee \
    --namespace $namespace

helm repo add jupyterhub https://hub.jupyter.org/helm-chart/
helm repo update

service_account="jupyter-k8s-pod-sa"
helm_name="jupyterhub-helm"
api="http://manatee.$namespace.svc.cluster.local"

helm upgrade --cleanup-on-fail \
    --set singleuser.image.name=docker.io/library/jupyterlab_manatee \
    --set singleuser.image.tag=latest \
    --set singleuser.image.pullPolicy=Never \
    --set singleuser.serviceAccountName=${service_account} \
    --set singleuser.extraEnv.DATA_CLEAN_ROOM_HOST=${api} \
    --set singleuser.extraEnv.DEPLOYMENT_ENV=${env} \
    --set singleuser.extraEnv.PROJECT_ID=${project_id} \
    --set singleuser.extraEnv.KEY_LOCALTION=${region} \
    --set singleuser.networkPolicy.enabled=false \
    --set singleuser.nodeSelector=null \
    --set prePuller.continuous.enabled=false \
    --set prePuller.hook.enabled=false \
    --install $helm_name jupyterhub/jupyterhub \
    --namespace ${namespace} \
    --version=3.0.3 \
    --values ../jupyterhub/config.yaml

echo "Deployment Completed."
echo "Try 'kubectl --namespace=$namespace get service proxy-public' to obtain external IP"
