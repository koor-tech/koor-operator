# koor-operator
An operator that installs Koor Storage Distro

## Description
This operator is equivalent to the following commands:

```sh
helm repo add koor-release https://charts.koor.tech/release
helm install --create-namespace --namespace <namespace> <namespace>-rook-ceph koor-release/rook-ceph -f utils/operatorValues.yaml
helm install --create-namespace --namespace <namespace> <namespace>-rook-ceph-cluster \
    --set operatorNamespace=<namespace> koor-release/rook-ceph-cluster -f utils/clusterValues.yaml
```

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [minikube](https://minikube.sigs.k8s.io/docs/start/) to get a local cluster for testing or run against a remote cluster.

**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e., whatever cluster `kubectl cluster-info` shows).

### Use a Local Docker Registry with Minikube
1. Make sure you start minikube with `--insecure-registry="localhost:5000"`. You might need to delete and restart minikube.

2. Enable the minikube [registry plugin](https://minikube.sigs.k8s.io/docs/handbook/registry/#docker-on-macos):

```sh
minikube addons enable registry
```

3. Redirect port 5000 on docker to port 5000 on the minikube

```sh
sudo docker run -d --network=host alpine/socat TCP-LISTEN:5000,reuseaddr,fork TCP:$(minikube ip):5000
```

1. Set the registry as `localhost:5000`

```sh
export REGISTRY_HOST=localhost:5000
```

## Run the operator
There are four ways to run the operator:

1. As a Go program outside a cluster
2. As a Deployment inside a Kubernetes cluster
3. Managed by the [Operator Lifecycle Manager (OLM)](https://sdk.operatorframework.io/docs/olm-integration/tutorial-bundle/#enabling-olm) in [bundle](https://sdk.operatorframework.io/docs/olm-integration/quickstart-bundle/) format
4. Using [helm](https://helm.sh/)

### Run locally outside the cluster
1. Generate certificates for local testing:

```sh
make local-certs
```

2. Install the CRDs into the cluster:

```sh
make install
```

3. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make local-certs install run`

#### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Running as a Deployment inside the cluster
1. Build and push your image to the registry. If `IMG` is not specified, it defaults to `$(REGISTRY_HOST)/koor-operator:v$(VERSION)`:

```sh
make docker-build docker-push
```

2. Install `cert-manager` if not already installed:

```sh
make cert-manager
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy
```

#### Undeploy controller
To undeploy the controller from the cluster:

```sh
make undeploy
```

#### Undeploy `cert-manager`
To undeploy `cert-manager` from the cluster:

```sh
make undeploy-cert-manager
```

### Deploy koor-operator with OLM
1. Make sure you have the `operator-sdk` binary [installed](https://sdk.operatorframework.io/docs/installation/), then install [OLM](https://sdk.operatorframework.io/docs/olm-integration/tutorial-bundle/#enabling-olm):

```sh
operator-sdk olm install
```

2. Build and push your image to the registry. If `IMG` is not specified, it defaults to `$(REGISTRY_HOST)/koor-operator:v$(VERSION)`:

```sh
make docker-build docker-push
```

3. Bundle the operator, then build and push the bundle image:

```sh
make bundle bundle-build bundle-push
```

4. Run the bundle:

```sh
operator-sdk run bundle <some registry>/koor-operator-bundle:v0.0.1
```

For example, using a local registry, the command becomes:

```sh
operator-sdk run bundle localhost:5000/koor-operator-bundle:v0.0.1 --use-http
```

### Install using Helm
1. Build and push your image to the registry. If `IMG` is not specified, it defaults to `$(REGISTRY_HOST)/koor-operator:v$(VERSION)`:

```sh
make docker-build docker-push
```

2. Install the helm chart to the cluster. Create a `values.yaml` file if necessary.

```sh
helm install koor-operator --namespace koor-operator --create-namespace charts/koor-operator
```

#### Uninstall helm chart
To undeploy the controller from the cluster:

```sh
helm uninstall koor-operator --namespace koor-operator
```

## Create the KoorCluster Custom Resource
When deploying the operator locally, as a deployment, or using the OLM, you need to create a KoorCluster custom resource. To do that, update the samples in `config/samples/...` to fit your needs, then create the Custom Resource:

```sh
kubectl apply -f config/samples/storage_v1alpha1_koorcluster.yaml
```

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the cluster reaches the desired state.

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2023 Koor Technologies, Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
