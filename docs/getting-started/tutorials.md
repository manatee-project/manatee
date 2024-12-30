# Tutorials

In this tutorial, we show the following scenario:

* Dataset: [US Health Insurance Dataset from Kaggle](https://www.kaggle.com/datasets/teertha/ushealthinsurancedataset/data)
* Task: Train a model predicting the insurance charge based on the data (use XGBoost regression)
* Data Provisioning: [MST](https://github.com/ryan112358/private-pgm/blob/9fc3a829f344a2ed0b662f4f8d92d5efe18e42ab/mechanisms/mst.py) (2018 NIST snythetic data challenge winner).

<iframe width="854" height="480" src="https://www.youtube.com/embed/Ig1NzZZ6yKE?si=-IjD6v2nqsrAuuuV" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>

The tutorial uses Google Confidential Space as the TEE backend, and uses preprocessed datasets.

## Overal Workflow

This tutorial uses a very simple prototype data SDK to interact with two different versions of datasets as shown below.

![ManaTEE Architecture](../assets/img/arch.png)
/// Caption
Overall ManaTEE Architecture and Demo Workflow
///


```
.
├── insurance.ipynb
├── regression.ipynb
├── data
│   ├── insurance.csv.s1
│   └── insurance.csv.s2
└── sdk
    ├── __init__.py
    ├── __version__.py
    └── data.py
```

## Prerequisite

gcloud CLI

## Data Provisioning

As mentioned earlier, data provisioning currently needs to be done manually by the data owner.
In summary, the data owner needs to:
(1) Use proper techniques to generate stage-1 dataset from the original (stage-2) dataset;
(2) Upload two different data versions to the cloud;
and (3) Use proper access control mechanism to both of the data versions.

We've already preprocessed stage-1 data using [NIST-MST](https://github.com/ryan112358/private-pgm/blob/9fc3a829f344a2ed0b662f4f8d92d5efe18e42ab/mechanisms/mst.py) differentially-private synthetic data generation. Thus, the tutorial starts from (2).

To provision data, first, prepare two different Google Cloud Storage buckets.
```
cd tutorials
```

Create two buckets (replace `<your bucket name>` with a unique string you'd like)

```
gcloud storage buckets create gs://<your bucket name>
```

Then, upload datasets into the bucket:

```
gcloud storage cp data/* gs://<your bucket name>
```

Now, let's use a proper access control mechanism to both of the datasets. We'll rely on a fine-grained access control provided by Google Cloud Storage.

In the demo, attestation is used only for checking the integrity of the workload and its output.
Thus, we omit encryption and attestation policy check based on image hash in this tutorial.
However, it does not mean that ManaTEE limits how the data provider should leverage the attestation policy.

<!-- In general, the data access control should be implemented based on remote attestation protocol provided by the TEE instance.
Google Confidential Space [recommends](https://cloud.google.com/docs/security/confidential-space) to use Google KMS and an OpenID Connect (OIDC) token provider.

 -->

## Data Permissions

### Stage-1 Permission

Give the JupyterLab single-user instance a permission to access stage-1 data.

```
gcloud storage buckets add-iam-policy-binding gs://<your bucket name>-stage-1 \
  --member=serviceAccount:jupyter-k8s-pod-sa@<GCP project ID>.iam.gserviceaccount.com \
  --role=roles/storage.objectViewer
```


### Stage-2 Permission

Now, let's give the TEE instance a permission to access stage-2 data.

First, create a service account that will be used by the TEE instance.

```
gcloud iam service-accounts create <your service account name>
```

Give the service account access to the bucket.

```
gcloud storage buckets add-iam-policy-binding gs://<your bucket name>-stage-2 \
  --member=serviceAccount:<your service account name>@<GCP project ID>.iam.gserviceaccount.com \
  --role=roles/storage.objectViewer
```

Now, we are going to use Google Confidential Space's token to make TEE impersonate the service account.

First, create a workload identity pool

```
gcloud iam workload-identity-pools create <your pool name> \
  --location=global
```

Then, grant the service account permission to impersonate the workload identity pool.
```
gcloud iam service-accounts add-iam-policy-binding \
    <your service account name>@<GCP project ID>.iam.gserviceaccount.com \
    --member="principalSet://iam.googleapis.com/projects/"$(gcloud projects describe <GCP project ID> \
        --format="value(projectNumber)")"/locations/global/workloadIdentityPools/<your pool name>/*" \
    --role=roles/iam.workloadIdentityUser
```

Finally, create a workload identity pool provider.

```
gcloud iam workload-identity-pools providers create-oidc attestation-verifier \
    --location=global \
    --workload-identity-pool=<your pool name> \
    --issuer-uri="https://confidentialcomputing.googleapis.com/" \
    --allowed-audiences="https://sts.googleapis.com" \
    --attribute-mapping="google.subject=assertion.sub" \
    --attribute-condition="assertion.swname == 'CONFIDENTIAL_SPACE' \
&& 'STABLE' in assertion.submods.confidential_space.support_attributes"
```

## Job Submission

### Prepare Jupyter Environment

Now, a JupyterHub user can use the notebook interface to write a script and submit it to the API via a JupyterLab extension called `jupyerlab-manatee`.
The extension is already installed in the deployed Jupyter Hub single-user image.

Please go to `tutorials` directory, and upload the `code` directory into the JupyterLab environment.

```
cd code
zip -r ../code.zip .
cd ..
```

Upload the `code.zip` to the JupyterLab interface.

Then, unzip it by running the following cell in the notebook

```
!unzip -o code.zip
```

### Stage-1: Programming

Open `insurance.ipynb`. Replace the following line with a proper bucket names.

```
bucket = sdk.init("<your stage-1 bucket name>", "<your stage-2 bucket name>")
```

When you run the cells, you will be able to see the stage-1 data getting fetched and processed. You can change the code to furthre explore the dataset.

### Stage-2: Secure Execution

When ready, the user can use the ManaTEE extension to submit a job to the ManaTEE API. The API address is determined at the deployment, and passed through as an environment variable.
