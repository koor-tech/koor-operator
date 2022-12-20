# koor-operator
An operator that installs Koor Storage Distro
## Description
This operator is equivalent to the following commands:

```sh
helm repo add koor-release https://charts.koor.tech/release
helm install --create-namespace --namespace koor-ceph koor-ceph koor-release/rook-ceph -f utils/operatorValues.yaml
helm install --create-namespace --namespace koor-ceph koor-ceph-cluster \
   --set operatorNamespace=koor-ceph koor-release/rook-ceph-cluster -f values-override.yaml
```

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [minikube](https://minikube.sigs.k8s.io/docs/start/) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Use a Local Docker Registry with Minikube
1. Enable the minikube [registry plugin](https://minikube.sigs.k8s.io/docs/handbook/registry/#docker-on-macos):
```sh
minikube addons enable registry
```

2. Redirect port 5000 on docker to port 5000 on the minikube
```sh
sudo docker run -d -p 5000:5000 alpine/socat TCP-LISTEN:5000,reuseaddr,fork TCP:$(minikube ip):5000
```

3. Use `localhost:5000` as `<some-registry>` in the commands below.

### Running on the cluster
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Install [cert-manager](https://cert-manager.io/docs/installation/) to enable webhooks:
```sh
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.10.1/cert-manager.yaml
```

3. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/koor-operator:tag
```

4. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/koor-operator:tag
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller to the cluster:

```sh
make undeploy
```

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/)
which provides a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster

### Test It Out
1. Generate certificates for local testing:

```sh
make generate-certs
```

2. Install the CRDs into the cluster:

```sh
make install
```

3. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make generate-certs install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
