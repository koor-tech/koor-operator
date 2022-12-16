# koor-operator
// TODO(user): Add simple overview of use/purpose

## Description
// TODO(user): An in-depth paragraph about your project and overview of use

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [minikube](https://minikube.sigs.k8s.io/docs/start/) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Use a Local Docker Registry with Minikube
1. Create a local docker registry:
```sh
sudo docker run -d -p 5000:5000 --restart=always --volume ~/.registry/storage:/var/lib/registry registry:2
```
2. Edit the `/etc/hosts` file on your development machine, adding the [hostname from minikube](https://minikube.sigs.k8s.io/docs/handbook/host-access/) `host.minikube.internal` on the same line as the entry for localhost.
3. Add the registry to insecure-registries in `/etc/docker/daemon.json`
```json
{
    "insecure-registries": [
        "localhost:5000"
    ]
}
```
3. Validate that the registry at `host.minikube.internal:5000` is reachable from your development machine.
```sh
$ curl host.minikube.internal:5000/v2/_catalog
{"repositories":[]}
```
4. Start your minikube cluster with the flag `--insecure-registry="host.minikube.internal:5000"`
4. Use `host.minikube.internal:5000` as `<some-registry>` in the commands below.

### Running on the cluster
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/koor-operator:tag
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

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
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster

### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

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
