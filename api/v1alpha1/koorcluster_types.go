/*
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
*/

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KoorClusterSpec defines the desired state of KoorCluster
type KoorClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Use all devices on nodes
	//+kubebuilder:default:=true
	UseAllDevices *bool `json:"useAllDevices,omitempty"`
	// Enable monitoring. Requires Prometheus to be pre-installed.
	//+kubebuilder:default:=true
	MonitoringEnabled *bool `json:"monitoringEnabled,omitempty"`
	// Enable the ceph dashboard for viewing cluster status
	//+kubebuilder:default:=true
	DashboardEnabled *bool `json:"dashboardEnabled,omitempty"`
	// Installs a debugging toolbox deployment
	//+kubebuilder:default:=true
	ToolboxEnabled *bool `json:"toolboxEnabled,omitempty"`
	// Specifies the upgrade options for new ceph versions
	UpgradeOptions UpgradeOptions `json:"upgradeOptions,omitempty"`
	// The name to use for KSD helm release.
	//+kubebuilder:default:=ksd
	KsdReleaseName string `json:"ksdReleaseName,omitempty"`
	// The name to use for KSD cluster helm release.
	//+kubebuilder:default:=ksd-cluster
	KsdClusterReleaseName string `json:"ksdClusterReleaseName,omitempty"`
}

// +kubebuilder:validation:Enum=disabled;notify;upgrade
type UpgradeMode string

const (
	UpgradeModeDisabled UpgradeMode = "disabled"
	UpgradeModeNotify   UpgradeMode = "notify"
	UpgradeModeUpgrade  UpgradeMode = "upgrade"
)

type UpgradeOptions struct {
	// Upgrade mode
	//+kubebuilder:default:=notify
	Mode UpgradeMode `json:"mode,omitempty"`
	// The api endpoint used to find the ceph latest version
	//+kubebuilder:default:="versions.koor.tech"
	Endpoint string `json:"endpoint,omitempty"`
	// The schedule to check for new versions. Uses CRON format as specified by https://github.com/robfig/cron/tree/v3.
	// Defaults to everyday at midnight in the local timezone.
	// To change the timezone, prefix the schedule with CRON_TZ=<Timezone>.
	// For example: "CRON_TZ=UTC 0 0 * * *" is midnight UTC.
	//+kubebuilder:default:="0 0 * * *"
	Schedule string `json:"schedule,omitempty"`
}

func (uo UpgradeOptions) IsEnabled() bool {
	return uo.Mode != UpgradeModeDisabled
}

// KoorClusterStatus defines the observed state of KoorCluster
type KoorClusterStatus struct {
	// The total resources available in the cluster nodes
	TotalResources Resources `json:"totalResources"`
	// Does the cluster meet the minimum recommended resources
	MeetsMinimumResources bool `json:"meetsMinimumResources"`
	// The current versions of rook and ceph
	CurrentVersions ProductVersions `json:"currentVersions,omitempty"`
	// The latest versions of rook and ceph
	LatestVersions *DetailedProductVersions `json:"latestVersions,omitempty"`
}

type ProductVersions struct {
	// The version of Kubernetes
	Kube string `json:"kube,omitempty"`
	// The version of the koor Operator
	KoorOperator string `json:"koorOperator,omitempty"`
	// The version of KSD
	Ksd string `json:"ksd,omitempty"`
	// The version of Ceph
	Ceph string `json:"ceph,omitempty"`
}

type DetailedProductVersions struct {
	// The detailed version of the koor Operator
	KoorOperator *DetailedVersion `json:"koorOperator,omitempty"`
	// The detailed version of KSD
	Ksd *DetailedVersion `json:"ksd,omitempty"`
	// The detailed version of Ceph
	Ceph *DetailedVersion `json:"ceph,omitempty"`
}

type DetailedVersion struct {
	Version        string `json:"version,omitempty"`
	ImageUri       string `json:"imageUri,omitempty"`
	ImageHash      string `json:"imageHash,omitempty"`
	HelmRepository string `json:"helmRepository,omitempty"`
	HelmChart      string `json:"helmChart,omitempty"`
}

type Resources struct {
	// The number of nodes in the cluster
	Nodes *resource.Quantity `json:"nodesCount,omitempty"`
	// Ephemeral Storage available
	Storage *resource.Quantity `json:"storage,omitempty"`
	// CPU cores available
	Cpu *resource.Quantity `json:"cpu,omitempty"`
	// Memory available
	Memory *resource.Quantity `json:"memory,omitempty"`
}

// Recommended Resources
var (
	minNodes   = resource.MustParse("4")
	minStorage = resource.MustParse("500G")
	minCpu     = resource.MustParse("19")
	minMemory  = resource.MustParse("44G")
)

func (r Resources) MeetsMinimum() bool {
	if r.Nodes.Cmp(minNodes) == -1 {
		return false
	}
	if r.Storage.Cmp(minStorage) == -1 {
		return false
	}
	if r.Cpu.Cmp(minCpu) == -1 {
		return false
	}
	if r.Memory.Cmp(minMemory) == -1 {
		return false
	}
	return true
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KoorCluster is the Schema for the koorclusters API
type KoorCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KoorClusterSpec   `json:"spec,omitempty"`
	Status KoorClusterStatus `json:"status,omitempty"`
}

func (k *KoorCluster) IsBeingDeleted() bool {
	return !k.ObjectMeta.DeletionTimestamp.IsZero()
}

const KoorClusterFinalizerName = "storage.koor.tech/finalizer"

//+kubebuilder:object:root=true

// KoorClusterList contains a list of KoorCluster
type KoorClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KoorCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KoorCluster{}, &KoorClusterList{})
}
