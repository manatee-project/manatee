# Building

ManaTEE uses [Bazel](https://bazel.build/install) for hermetic builds.
Bazel is aware of all required tools and dependencies, thus building images is as easy as:

```
bazel build //...
```

Find individual rules from corresponding `BUILD.bazel` files.

## Components

`app` directory contains the source codes of the data clean room which has three components:

* `dcr_tee` contains tools that are used in the base image of stage2 such as a tool generates custom attestation report within GCP confidential space.
* `dcr_api` is the backend service of the data clean room that processes the request from jupyterlab. 
* `dcr_monitor` is a cron job that monitors the execution of each job. The monitor is deployed to Kubernetes cluster and scheduled to run every minute.
* `jupyterlab_manatee` is an JupyterLab extension for data clean room that submits a job on the fronted and queries the status of the jobs.

## Loading Container Images

If you'd like to load the images in your local container runtime (e.g., Docker), you can use `oci_load` rules.

```shell
bazel query 'kind("oci_load", "//app/...")' | xargs -n1 bazel run
```

# Testing

To run all tests, run:

```
bazel test //...
```