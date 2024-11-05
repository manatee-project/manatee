#!/bin/bash

# NOTE: this is temporary solution for making config file for the app.
# Ideally, the config should be provisioned at the deployment, not at the build.
# Right now, we just generate one so that we can test the app.

GCP_PROJECT_ID="$1"
DEPLOY_ENV="$2"
GCP_REGION="$3"
GCP_ZONE="$4"

cat <<EOF | envsubst
CloudProvider:
  GCP:
    Project: "${GCP_PROJECT_ID}"
    HubBucket: "dcr-${DEPLOY_ENV}-hub"
    Zone: "${GCP_ZONE}"
    Region: "${GCP_REGION}"
    Debug: false
    Env: ${DEPLOY_ENV}
Cluster:
  PodServiceAccount: "dcr-k8s-pod-sa"
EOF
