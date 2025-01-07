#!/bin/sh
STAGE_1_BUCKET=dayeol-manatee-tutorial-stage-1
STAGE_2_BUCKET=dayeol-manatee-tutorial-stage-2
WORKLOAD_IDENTITY_POOL_NAME=manatee-cvm-pool
TEE_SERVICE_ACCOUNT=dayeol-tee-sa

VAR_FILE="../env.bzl"
if [ ! -f "$VAR_FILE" ]; then
    echo "Error: Variables file does not exist."
    exit 1
fi

VAR_FILE=$(realpath $VAR_FILE)
source $VAR_FILE

# data provisioning

gcloud storage buckets create gs://$STAGE_1_BUCKET
gcloud storage buckets create gs://$STAGE_2_BUCKET

gcloud storage cp data/stage1/insurance.csv gs://$STAGE_1_BUCKET
gcloud storage cp data/stage2/insurance.csv gs://$STAGE_2_BUCKET

# data permissions: stage 1

gcloud storage buckets add-iam-policy-binding gs://$STAGE_1_BUCKET \
  --member=serviceAccount:jupyter-$env-pod-sa@$project_id.iam.gserviceaccount.com \
  --role=roles/storage.objectViewer 

# data permissions: stage 2

gcloud iam service-accounts create $TEE_SERVICE_ACCOUNT

gcloud storage buckets add-iam-policy-binding gs://$STAGE_2_BUCKET \
  --member=serviceAccount:$TEE_SERVICE_ACCOUNT@$project_id.iam.gserviceaccount.com \
  --role=roles/storage.objectViewer

gcloud iam workload-identity-pools create $WORKLOAD_IDENTITY_POOL_NAME \
  --location=global

gcloud iam service-accounts add-iam-policy-binding \
    $TEE_SERVICE_ACCOUNT@$project_id.iam.gserviceaccount.com \
    --member="principalSet://iam.googleapis.com/projects/"$(gcloud projects describe $project_id \
        --format="value(projectNumber)")"/locations/global/workloadIdentityPools/$WORKLOAD_IDENTITY_POOL_NAME/*" \
    --role=roles/iam.workloadIdentityUser

gcloud iam workload-identity-pools providers create-oidc attestation-verifier \
    --location=global \
    --workload-identity-pool=$WORKLOAD_IDENTITY_POOL_NAME \
    --issuer-uri="https://confidentialcomputing.googleapis.com/" \
    --allowed-audiences="https://sts.googleapis.com" \
    --attribute-mapping="google.subject=\"gcpcs::\"+assertion.submods.container.image_digest+\"::\"+assertion.submods.gce.project_number+\"::\"+assertion.submods.gce.instance_id" \
    --attribute-condition="assertion.swname == 'CONFIDENTIAL_SPACE' && 'STABLE' in assertion.submods.confidential_space.support_attributes"
