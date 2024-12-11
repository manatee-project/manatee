# Minikube Deployment
Minikube deployment supports users to experience most functionality of manatee withou a google account. 

1. Install `minikube` cli.
2. Create a `minikube` cluster. `minikube start --memory=12192mb --cpus=8 --disk-size=50g --insecure-registry "10.0.0.0/24" && minikube addons enable registry`
3. Create minimal resources in the minikube cluster. `pushd resources/minikube && ./apply.sh && popd `
4. Build the imaeg `eval $(minikube docker-env) && bazel build //... && bazel run app/dcr_api/load_image && bazel run app/dcr_monitor/load_image && bazel run app/dcr_tee/load_image && bazel run app/jupyterlab_manatee/load_image`
5. Deploy the manatee to minikube. `pushd deployment/minikube && ./deploy.sh && popd`
6. you can port-forward traffic to the k8s Service proxy-public with kubectl to access it from your computer. `kubectl --namespace=manatee port-forward service/proxy-public 8080:http`. Try insecure HTTP access: http://localhost