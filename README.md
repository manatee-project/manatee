<img width=100% src="logo.png">

# ManaTEE Project

> Note: we are releasing an alpha version, which may miss some necessary features. 

ManaTEE is an open-source project for easily building and deploying data collaboration framework to the cloud using trusted execution environments (TEEs).
It allows users to easily collaborate on private datasets without leaking privacy of individual data.
ManaTEE achieves this by combining different privacy-enhancing technologies (PETs) in different programming stages.

In summary, ManaTEE is great tool for data collaboration with the following features.

* **Interactive Programming**: ManaTEE integrates with an existing Jupyter Notebook interface such that the data analysts can program interactively with popular languages like Python
* **Multiparty**: ManaTEE allows multi-party data collaboration without needing to send the private data to each other
* **Cloud-Ready**: ManaTEE can be easily deployed in TEEs in the cloud, including Google Confidential Space
* **Accurate Results**: ManaTEE does not sacrifice accuracy for data privacy. This is achieved by a two-stage approach with different PETs applied to each stage.

### What is Different from Other Data Collaboration Frameworks?

Data collaboration is not a new concept, and numerous data collaboration frameworks already exist.
However, different frameworks try to apply different privacy-enhancing technologies (PETs), which have different strengths and weaknesses.
ManaTEE tries to utilize different PETs in different programming stages to maximize the usability while protecting individual data privacy.
Specifically, ManaTEE divides data analytics in two stages: *Programming Stage* and *Secure Execution Stage*.

![Alt text](two-stage.png)

In Programming Stage, the data scientist uses Jupyter Notebook interface to explore the general data structure and statistical characteristics. 
The data providers can determine how they protect privacy of their data. 
For example, they can use differentially-private synthetic data, completely random data, or partial public data.
This means that the it mathematically limits the leakage of privacy of individual data records.
The finished notebook files can then be submitted to the Secure Execution Stage.

In Secure Execution Stage, the submitted notebook file is built into an image, and scheduled to a confidential virtual machine (CVMs) in the cloud.
The data providers can set up their data such that only *attested* program can fetch the data. 
By using attestation, the data providers can control which program can access their data. 
TEE also assures the data scientists that the integrity of their program and the legitimacy of the output from executions by providing JWT-based attestation report.

### Use Cases

There are many potential use cases of the ManaTEE:

* **Trusted Research Environments (TREs)**: Some data may be valuable to various research on public health, economic impact, and many other fields.
TREs are a secure environment where authorized/vetted researchers and organizations can access the data. The data provider can choose to use ManaTEE to build their TRE.
Currently, [TikTok's Research Tools Virtual Compute Environment (VCE)](https://developers.tiktok.com/doc/vce-getting-started) is built on top of ManaTEE.

* **Advertisement and Marketing**: Ads is a popular use case of data collaboration frameworks. ManaTEE can be used for [lookalike segment analysis](https://en.wikipedia.org/wiki/Lookalike_audience) for advertisers, or [Ad Tracking](https://en.wikipedia.org/wiki/Ad_tracking) with private user data.

* **Machine Learning**: ManaTEE can be useful for machine learning involving private data or models. For example, a private model provider can provide their model for fine-tuning, but do not reveal the actual model in the Programming Stage.

### Project Status

We are releasing an alpha version, which may miss some necessary features.

|                         | Current (Alpha)          | Future                    |
|-------------------------|--------------------------|---------------------------|
| **Users**               | One-Way Collaboration    | Multi-Way Collaboration   |
| **Backend**             | Single Backend (Goole Cloud Platform)     | Multiple Backend          |
| **Data Provisioning**   | Manual                   | Automated                 |
| **Policy and Attestation** | Manual                | Automated                 |
| **Compute**             | CPU                      | CPU/GPU                   |

# Getting Started

## Prerequisites
* A valid GCP account that has ability to create/destroy resources. For a GCP project, please enable the following apis:
    - serviceusage.googleapis.com
    - compute.googleapis.com
    - container.googleapis.com
    - cloudkms.googleapis.com
    - servicenetworking.googleapis.com
    - cloudresourcemanager.googleapis.com
    - sqladmin.googleapis.com
    - confidentialcomputing.googleapis.com
* [Gcloud CLI](https://cloud.google.com/sdk/docs/install) Login to the GCP `gcloud auth login && gcloud auth application-default login && gcloud components install gke-gcloud-auth-plugin`
* [Terraform](https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli) Terraform is an infrastructure as code tool that enables you to safely and predictably provision and manage infrastructure in any cloud.
* [Helm](https://helm.sh/docs/intro/install/) Helm is a package manager for Kubernetes that allows developers and operators to more easily package, configure, and deploy applications and services onto Kubernetes clusters.
* [Hertz](https://github.com/cloudwego/hertz) Hertz is a high-performance, high-usability, extensible HTTP framework for Go. It’s designed to make it easy for developers to build microservices.

## Deploying in Google Cloud Platform (GCP)
### Defining environment variables
First, copy the example environment variables template to the existing directory.
```
cp .env.example env.bzl
```
Edit the variables in `env.bzl`. The `env.bzl` file is the one that really takes effect, the other files are just templates. The double quotes around a variable name are needed. For example:
```
env="dev"                        # the deployment environment
project_id="you project id"      # gcp project id
region=""                        # the region that the resources created in
zone=""                          # the zone that the resources created in
```

### Preparing resources
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

### Building and Pushing Images
`app` directory contains the source codes of the data clean room which has three components:

* `dcr_tee` contains tools that are used in the base image of stage2 such as a tool generates custom attestation report within GCP confidential space.
* `dcr_api` is the backend service of the data clean room that processes the request from jupyterlab. 
* `dcr_monitor` is a cron job that monitors the execution of each job. The monitor is deployed to Kubernetes cluster and scheduled to run every minute.
* `jupyterlab_manatee` is an JupyterLab extension for data clean room that submits a job on the fronted and queries the status of the jobs.

[Bazel](https://bazel.build/install) is required to build all of the binaries and push them to the artifact registry.

```shell 
gcloud auth configure-docker us-docker.pkg.dev # authenticate to artifact registry
bazel run //:push_all_images --action_env=namespace=<namespace-to-deploy>
```

> [!IMPORTANT]
> the `--action_env=namespace=<namespace-to-deploy>` flag is required.

You can also push images separately by this command. Replace `<app>` by the directory name under `/app` (e.g., dcr_api)

```
bazel run //:push_<app>_image --action_env=namespace=<namespace-to-deploy>
```


If you'd like to load the images in your local container runtime (e.g., Docker), you can use `oci_load` rules.

```shell
bazel query 'kind("oci_load", "//app/...")' | xargs -n1 bazel run
```

Find individual rules from corresponding `BUILD.bazel` files.

### Deploying 

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
