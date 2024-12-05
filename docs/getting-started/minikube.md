# Minikube Deployment
Minikube deployment supports users to experience most functionality of manatee withou a google account. 

1. Install `minikube` cli.
2. Create a `minikube` cluster. `minikube start --memory=8192mb --cpus=8 --disk-size=50g`
3. Create minimal resources in the minikube cluster. `pushd resources/minikube && ./apply.sh && popd `
4. Build the imaeg `bazel build //... && bazel run app/dcr_api/load_image && bazel run app/dcr_monitor/load_image && bazel run app/dcr_tee/load_image && bazel run app/jupyterlab_manatee/load_image`
5. Deploy the manatee to minikube. `pushd deployment/minikube && ./deploy.sh`
6. 