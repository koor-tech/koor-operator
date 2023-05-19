---
title: Ceph Operator Helm Chart
---
{{ template "generatedDocsWarning" . }}

Installs [Koor Operator](https://github.com/koor-tech/koor-operator) to create, configure, and manage Koor Storage Distribution on Kubernetes.

## Introduction

This chart bootstraps a [Koor Operator](https://github.com/koor-tech/koor-operator) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites

* Kubernetes 1.19+
* Helm 3.x

See the [Helm support matrix](https://helm.sh/docs/topics/version_skew/) for more details.

## Installing

The Ceph Operator helm chart will install the basic components necessary to create a storage platform for your Kubernetes cluster.

1. Add the Koor Helm repo
2. Install the Helm chart
3. [Create a Koor Storage cluster](https://docs.koor.tech/v1.11/Getting-Started/quickstart/#create-a-ceph-cluster).

The `helm install` command deploys the Koor Operator on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation. It is recommended that the Koor Operator be installed into the `koor-operator` namespace (you will install your clusters into separate namespaces).

```console
helm repo add koor-operator https://koor-tech.github.io/koor-operator
helm install --create-namespace --namespace koor-operator koor-operator koor-operator/koor-operator -f values.yaml
```

For example settings, see the next section or [values.yaml](/charts/koor-operator/values.yaml).

## Configuration

The following table lists the configurable parameters of the rook-operator chart and their default values.

{{ template "chart.valuesTable" . }}

## Uninstalling the Chart

To see the currently installed Rook chart:

```console
helm ls --namespace koor-operator
```

To uninstall/delete the `koor-operator` deployment:

```console
helm delete --namespace koor-operator koor-operator
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

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
