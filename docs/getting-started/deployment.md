# GCP Deployment

## Prerequisites

Currently, ManaTEE requires Google Cloud Platform (GCP) for deployment, as it requires cloud-provided TEE.
In the future, we will support more cloud backends as well as local test deployment (See [#31](https://github.com/manatee-project/manatee/issues/31)).

Because of the cloud resource requirement, we recommend a cloud admin to create all the resources by following the steps below.

### Cloud Setup

* A valid GCP account that has ability to create/destroy resources. For a GCP project, please enable the following apis:
    - serviceusage.googleapis.com
    - compute.googleapis.com
    - container.googleapis.com
    - cloudkms.googleapis.com
    - servicenetworking.googleapis.com
    - cloudresourcemanager.googleapis.com
    - sqladmin.googleapis.com
    - confidentialcomputing.googleapis.com

### Tools
* [Gcloud CLI](https://cloud.google.com/sdk/docs/install) Login to the GCP `gcloud auth login && gcloud auth application-default login && gcloud components install gke-gcloud-auth-plugin`
* [Terraform](https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli) Terraform is an infrastructure as code tool that enables you to safely and predictably provision and manage infrastructure in any cloud.
* [Helm](https://helm.sh/docs/intro/install/) Helm is a package manager for Kubernetes that allows developers and operators to more easily package, configure, and deploy applications and services onto Kubernetes clusters.
* [Hertz](https://github.com/cloudwego/hertz) Hertz is a high-performance, high-usability, extensible HTTP framework for Go. It’s designed to make it easy for developers to build microservices.

## Create Resources

The resources are created and managed by the project administrator who has the `Owner` role in the GCP project. Make sure you have correctly defined environment variables in the `env.bzl`. Only the project administrator is responsible to run these commands to create resources.

`resources/global` directory contains the global resources including: clusters, cloud sql instance, database, docker repositories, and service accounts. These resource are global and only created once.
```
pushd resources/global
./apply.sh
popd
```

`resources/deployment` directory includes the resources releated to kunernates including: kubernetes namespace, role, secret. These resources are created under different namespace. So the namespace parameter is required, and you can create different deployments under different namespaces.
```shell 
pushd resources/deployment
./apply.sh --namespace=<namespace-to-deploy>
popd
```

## Pushing Images

```shell 
gcloud auth configure-docker us-docker.pkg.dev # authenticate to artifact registry
bazel run //:push_all_images --action_env=namespace=<namespace-to-deploy>
```

> [!IMPORTANT]
> the `--action_env=namespace=<namespace-to-deploy>` flag is required.

You can also push images separately by this command. Replace `<app>` by the directory name under `/app` (e.g., api)

```
bazel run //:push_<app>_image --action_env=namespace=<namespace-to-deploy>
```

## Deploying in Google Cloud Platform (GCP)

### Defining environment variables
First, copy the example environment variables template to the existing directory.
```
cp .env.example env.bzl
```
Edit the variables in `env.bzl`. The `env.bzl` file is the one that really takes effect, the other files are just templates. The double quotes around a variable name are needed. For example:

``` sh title="env.bzl"
env="dev"                        # the deployment environment
project_id="you project id"      # gcp project id
region=""                        # the region that the resources created in
zone=""                          # the zone that the resources created in
```

### Deploy

Deploy data clean room and jupyterhub by helm chart.
```shell 
source env.bzl
gcloud container clusters get-credentials dcr-$env-cluster --zone $zone --project $project_id

pushd deployment
./deploy.sh --namespace=<namespace-to-deploy>
popd
```
When deployment is complete, you can follow the output of the script to get the public ip of jupyterhub. 
```
kubectl --namespace=<namespace-to-deploy> get service proxy-public
```
