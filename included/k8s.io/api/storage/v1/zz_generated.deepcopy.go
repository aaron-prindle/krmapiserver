// +build !ignore_autogenerated

/*
Copyright The Kubernetes Authors.

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1

import (
	corev1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/core/v1"
	runtime "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StorageClass) DeepCopyInto(out *StorageClass) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.Parameters != nil {
		in, out := &in.Parameters, &out.Parameters
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.ReclaimPolicy != nil {
		in, out := &in.ReclaimPolicy, &out.ReclaimPolicy
		*out = new(corev1.PersistentVolumeReclaimPolicy)
		**out = **in
	}
	if in.MountOptions != nil {
		in, out := &in.MountOptions, &out.MountOptions
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.AllowVolumeExpansion != nil {
		in, out := &in.AllowVolumeExpansion, &out.AllowVolumeExpansion
		*out = new(bool)
		**out = **in
	}
	if in.VolumeBindingMode != nil {
		in, out := &in.VolumeBindingMode, &out.VolumeBindingMode
		*out = new(VolumeBindingMode)
		**out = **in
	}
	if in.AllowedTopologies != nil {
		in, out := &in.AllowedTopologies, &out.AllowedTopologies
		*out = make([]corev1.TopologySelectorTerm, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StorageClass.
func (in *StorageClass) DeepCopy() *StorageClass {
	if in == nil {
		return nil
	}
	out := new(StorageClass)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StorageClass) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StorageClassList) DeepCopyInto(out *StorageClassList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]StorageClass, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StorageClassList.
func (in *StorageClassList) DeepCopy() *StorageClassList {
	if in == nil {
		return nil
	}
	out := new(StorageClassList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StorageClassList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeAttachment) DeepCopyInto(out *VolumeAttachment) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeAttachment.
func (in *VolumeAttachment) DeepCopy() *VolumeAttachment {
	if in == nil {
		return nil
	}
	out := new(VolumeAttachment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VolumeAttachment) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeAttachmentList) DeepCopyInto(out *VolumeAttachmentList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VolumeAttachment, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeAttachmentList.
func (in *VolumeAttachmentList) DeepCopy() *VolumeAttachmentList {
	if in == nil {
		return nil
	}
	out := new(VolumeAttachmentList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VolumeAttachmentList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeAttachmentSource) DeepCopyInto(out *VolumeAttachmentSource) {
	*out = *in
	if in.PersistentVolumeName != nil {
		in, out := &in.PersistentVolumeName, &out.PersistentVolumeName
		*out = new(string)
		**out = **in
	}
	if in.InlineVolumeSpec != nil {
		in, out := &in.InlineVolumeSpec, &out.InlineVolumeSpec
		*out = new(corev1.PersistentVolumeSpec)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeAttachmentSource.
func (in *VolumeAttachmentSource) DeepCopy() *VolumeAttachmentSource {
	if in == nil {
		return nil
	}
	out := new(VolumeAttachmentSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeAttachmentSpec) DeepCopyInto(out *VolumeAttachmentSpec) {
	*out = *in
	in.Source.DeepCopyInto(&out.Source)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeAttachmentSpec.
func (in *VolumeAttachmentSpec) DeepCopy() *VolumeAttachmentSpec {
	if in == nil {
		return nil
	}
	out := new(VolumeAttachmentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeAttachmentStatus) DeepCopyInto(out *VolumeAttachmentStatus) {
	*out = *in
	if in.AttachmentMetadata != nil {
		in, out := &in.AttachmentMetadata, &out.AttachmentMetadata
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.AttachError != nil {
		in, out := &in.AttachError, &out.AttachError
		*out = new(VolumeError)
		(*in).DeepCopyInto(*out)
	}
	if in.DetachError != nil {
		in, out := &in.DetachError, &out.DetachError
		*out = new(VolumeError)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeAttachmentStatus.
func (in *VolumeAttachmentStatus) DeepCopy() *VolumeAttachmentStatus {
	if in == nil {
		return nil
	}
	out := new(VolumeAttachmentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeError) DeepCopyInto(out *VolumeError) {
	*out = *in
	in.Time.DeepCopyInto(&out.Time)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeError.
func (in *VolumeError) DeepCopy() *VolumeError {
	if in == nil {
		return nil
	}
	out := new(VolumeError)
	in.DeepCopyInto(out)
	return out
}
