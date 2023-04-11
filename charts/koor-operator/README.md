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
