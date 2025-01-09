# Test Deployment on Minikube

We also made it possible to deploy and test ManaTEE without having any cloud account.
Our test deployment uses a local Minikube cluster with a few components that replaces cloud resources.
With this, users can quickly test and try ManaTEE JupyterLab extension and the API without having an actual TEE backend.

## Prerequisite

First, Install [Minikube CLI](https://minikube.sigs.k8s.io/docs/start/).

Then, create a minikube cluster with enough memory. We need larger memory because of the Kaniko jobs.

```
minikube start --memory=12192mb --cpus=8 --disk-size=50g --insecure-registry "10.0.0.0/24"
```

## Create Cluster Resources

Once minikube cluster is up and running, create the resources in the minikube cluster

```
pushd resources/minikube
./apply.sh
popd
```

## Build Images

Now, build the images and load it into the Docker.
Minikube has its own Docker engine running inside the cluster.
Thus, we first need to point the local Docker client to the Docker engine inside minikube 

```
eval $(minikube docker-env)
```

Then, run the following command to load all images

```
bazelisk run //:load_all_images
```

## Setup Registry

The API requires artifact registry to store the TEE base image.
Thus, we use minikube's registry addon to host the image.

Enable the registry
```
minikube addons enable registry
```

RUN a proxy to connect to minikube registry and push dcr_tee image to minikube registry.
```
docker run --rm -it --network=host alpine ash -c "apk add socat && socat TCP-LISTEN:5000,reuseaddr,fork TCP:$(minikube ip):5000"
```

Open another terminal, and run

```
eval $(minikube docker-env)
docker tag dcr_tee localhost:5000/dcr_tee && docker push localhost:5000/dcr_tee
```

You can close the proxy after the docker push.

## Deploy

Now, you can deploy ManaTEE to minikube.

```
pushd deployment/minikube
./deploy.sh
popd
```

## Accessing JupyterHub

You can port-forward traffic to the k8s Service proxy-public with kubectl to access it from your computer. `kubectl --namespace=manatee port-forward service/proxy-public 8080:http`. 

Try insecure HTTP access: http://localhost:8080
