//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KoorCluster) DeepCopyInto(out *KoorCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KoorCluster.
func (in *KoorCluster) DeepCopy() *KoorCluster {
	if in == nil {
		return nil
	}
	out := new(KoorCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KoorCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KoorClusterList) DeepCopyInto(out *KoorClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]KoorCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KoorClusterList.
func (in *KoorClusterList) DeepCopy() *KoorClusterList {
	if in == nil {
		return nil
	}
	out := new(KoorClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KoorClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KoorClusterSpec) DeepCopyInto(out *KoorClusterSpec) {
	*out = *in
	if in.UseAllDevices != nil {
		in, out := &in.UseAllDevices, &out.UseAllDevices
		*out = new(bool)
		**out = **in
	}
	if in.MonitoringEnabled != nil {
		in, out := &in.MonitoringEnabled, &out.MonitoringEnabled
		*out = new(bool)
		**out = **in
	}
	if in.DashboardEnabled != nil {
		in, out := &in.DashboardEnabled, &out.DashboardEnabled
		*out = new(bool)
		**out = **in
	}
	if in.ToolboxEnabled != nil {
		in, out := &in.ToolboxEnabled, &out.ToolboxEnabled
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KoorClusterSpec.
func (in *KoorClusterSpec) DeepCopy() *KoorClusterSpec {
	if in == nil {
		return nil
	}
	out := new(KoorClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KoorClusterStatus) DeepCopyInto(out *KoorClusterStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KoorClusterStatus.
func (in *KoorClusterStatus) DeepCopy() *KoorClusterStatus {
	if in == nil {
		return nil
	}
	out := new(KoorClusterStatus)
	in.DeepCopyInto(out)
	return out
}
