/*
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
*/

package v1alpha1

import (
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
	DahsboardEnabled *bool `json:"dashboardEnabled,omitempty"`
	// Installs a debugging toolbox deployment
	ToolboxEnabled *bool `json:"toolboxEnabled,omitempty"`
}

// KoorClusterStatus defines the observed state of KoorCluster
type KoorClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
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
