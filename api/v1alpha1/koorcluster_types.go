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
	UseAllDevices *bool `json:"useAllDevices,omitempty"`
	// Enable monitoring. Requires Prometheus to be pre-installed.
	MonitoringEnabled *bool `json:"monitoringEnabled,omitempty"`
	// Enable the ceph dashboard for viewing cluster status
	DashboardEnabled *bool `json:"dashboardEnabled,omitempty"`
	// Installs a debugging toolbox deployment
	ToolboxEnabled *bool `json:"toolboxEnabled,omitempty"`
	// Specifies the notification options for new ceph versions
	NotificationOptions NotificationOptions `json:"notificationOptions,omitempty"`
}

type NotificationOptions struct {
	Enabled bool `json:"enabled,omitempty"`
	// The repository where the ceph image is pulled. Defaults to quai.io
	CephImageRepository string `json:"cephImageRepository,omitempty"`
	// The schedule to check for new versions. Uses CRON format. Defaults to everyday at midnight
	Schedule string `json:"schedule,omitempty"`
}

// KoorClusterStatus defines the observed state of KoorCluster
type KoorClusterStatus struct {
	// The total resources available in the cluster nodes
	TotalResources Resources `json:"totalResources"`
	// Does the cluster meet the minimum recommended resources
	MeetsMinimumResources bool `json:"meetsMinimumResources"`
	// The current versions of rook and ceph
	CurrentVersions Versions `json:"currentVersions,omitempty"`
	// The latest versions of rook and ceph
	LatestVersions Versions `json:"latestVersions,omitempty"`
}

type Versions struct {
	// The version of Ceph
	Ceph string `json:"ceph,omitempty"`
	// The version of Rook
	Rook string `json:"rook,omitempty"`
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
