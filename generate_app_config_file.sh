#!/bin/bash

# NOTE: this is temporary solution for making config file for the app.
# Ideally, the config should be provisioned at the deployment, not at the build.
# Right now, we just generate one so that we can test the app.

GCP_PROJECT_ID="$1"
GCP_PROJECT_NUMBER="$2"
DEPLOY_ENV="$3"
GCP_REGION="$4"
GCP_ZONE="$5"

cat <<EOF | envsubst
CloudProvider:
  GCP:
    Project: "${GCP_PROJECT_ID}"
    ProjectNumber: "${GCP_PROJECT_NUMBER}"
    Repository: "dcr-${DEPLOY_ENV}-user-images"
    HubBucket: "dcr-${DEPLOY_ENV}-hub"
    CvmServiceAccount: "dcr-${DEPLOY_ENV}-cvm-sa"
    Zone: "${GCP_ZONE}"
    Region: "${GCP_REGION}"
    CPUs: 2
    DiskSize: 50
    Debug: false
    KeyRing: "dcr-${DEPLOY_ENV}-keyring"
    WorkloadIdentityPool: "dcr-${DEPLOY_ENV}-pool"
    IssuerUri: "https://confidentialcomputing.googleapis.com/"
    AllowedAudiences: ["https://sts.googleapis.com"]
    Network: "dcr-${DEPLOY_ENV}-network"
    Subnetwork: "dcr-${DEPLOY_ENV}-subnetwork"
    Env: ${DEPLOY_ENV}
Cluster:
  PodServiceAccount: "dcr-k8s-pod-sa"
EOF
