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

// Code generated by conversion-gen. DO NOT EDIT.

package v1alpha1

import (
	unsafe "unsafe"

	corev1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/core/v1"
	v1alpha1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/storage/v1alpha1"
	conversion "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/conversion"
	runtime "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	core "github.com/aaron-prindle/krmapiserver/pkg/apis/core"
	v1 "github.com/aaron-prindle/krmapiserver/pkg/apis/core/v1"
	storage "github.com/aaron-prindle/krmapiserver/pkg/apis/storage"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// RegisterConversions adds conversion functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterConversions(s *runtime.Scheme) error {
	if err := s.AddGeneratedConversionFunc((*v1alpha1.VolumeAttachment)(nil), (*storage.VolumeAttachment)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_VolumeAttachment_To_storage_VolumeAttachment(a.(*v1alpha1.VolumeAttachment), b.(*storage.VolumeAttachment), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*storage.VolumeAttachment)(nil), (*v1alpha1.VolumeAttachment)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_storage_VolumeAttachment_To_v1alpha1_VolumeAttachment(a.(*storage.VolumeAttachment), b.(*v1alpha1.VolumeAttachment), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v1alpha1.VolumeAttachmentList)(nil), (*storage.VolumeAttachmentList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_VolumeAttachmentList_To_storage_VolumeAttachmentList(a.(*v1alpha1.VolumeAttachmentList), b.(*storage.VolumeAttachmentList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*storage.VolumeAttachmentList)(nil), (*v1alpha1.VolumeAttachmentList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_storage_VolumeAttachmentList_To_v1alpha1_VolumeAttachmentList(a.(*storage.VolumeAttachmentList), b.(*v1alpha1.VolumeAttachmentList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v1alpha1.VolumeAttachmentSource)(nil), (*storage.VolumeAttachmentSource)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_VolumeAttachmentSource_To_storage_VolumeAttachmentSource(a.(*v1alpha1.VolumeAttachmentSource), b.(*storage.VolumeAttachmentSource), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*storage.VolumeAttachmentSource)(nil), (*v1alpha1.VolumeAttachmentSource)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_storage_VolumeAttachmentSource_To_v1alpha1_VolumeAttachmentSource(a.(*storage.VolumeAttachmentSource), b.(*v1alpha1.VolumeAttachmentSource), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v1alpha1.VolumeAttachmentSpec)(nil), (*storage.VolumeAttachmentSpec)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_VolumeAttachmentSpec_To_storage_VolumeAttachmentSpec(a.(*v1alpha1.VolumeAttachmentSpec), b.(*storage.VolumeAttachmentSpec), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*storage.VolumeAttachmentSpec)(nil), (*v1alpha1.VolumeAttachmentSpec)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_storage_VolumeAttachmentSpec_To_v1alpha1_VolumeAttachmentSpec(a.(*storage.VolumeAttachmentSpec), b.(*v1alpha1.VolumeAttachmentSpec), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v1alpha1.VolumeAttachmentStatus)(nil), (*storage.VolumeAttachmentStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_VolumeAttachmentStatus_To_storage_VolumeAttachmentStatus(a.(*v1alpha1.VolumeAttachmentStatus), b.(*storage.VolumeAttachmentStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*storage.VolumeAttachmentStatus)(nil), (*v1alpha1.VolumeAttachmentStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_storage_VolumeAttachmentStatus_To_v1alpha1_VolumeAttachmentStatus(a.(*storage.VolumeAttachmentStatus), b.(*v1alpha1.VolumeAttachmentStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v1alpha1.VolumeError)(nil), (*storage.VolumeError)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_VolumeError_To_storage_VolumeError(a.(*v1alpha1.VolumeError), b.(*storage.VolumeError), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*storage.VolumeError)(nil), (*v1alpha1.VolumeError)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_storage_VolumeError_To_v1alpha1_VolumeError(a.(*storage.VolumeError), b.(*v1alpha1.VolumeError), scope)
	}); err != nil {
		return err
	}
	return nil
}

func autoConvert_v1alpha1_VolumeAttachment_To_storage_VolumeAttachment(in *v1alpha1.VolumeAttachment, out *storage.VolumeAttachment, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	if err := Convert_v1alpha1_VolumeAttachmentSpec_To_storage_VolumeAttachmentSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_v1alpha1_VolumeAttachmentStatus_To_storage_VolumeAttachmentStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

// Convert_v1alpha1_VolumeAttachment_To_storage_VolumeAttachment is an autogenerated conversion function.
func Convert_v1alpha1_VolumeAttachment_To_storage_VolumeAttachment(in *v1alpha1.VolumeAttachment, out *storage.VolumeAttachment, s conversion.Scope) error {
	return autoConvert_v1alpha1_VolumeAttachment_To_storage_VolumeAttachment(in, out, s)
}

func autoConvert_storage_VolumeAttachment_To_v1alpha1_VolumeAttachment(in *storage.VolumeAttachment, out *v1alpha1.VolumeAttachment, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	if err := Convert_storage_VolumeAttachmentSpec_To_v1alpha1_VolumeAttachmentSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_storage_VolumeAttachmentStatus_To_v1alpha1_VolumeAttachmentStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

// Convert_storage_VolumeAttachment_To_v1alpha1_VolumeAttachment is an autogenerated conversion function.
func Convert_storage_VolumeAttachment_To_v1alpha1_VolumeAttachment(in *storage.VolumeAttachment, out *v1alpha1.VolumeAttachment, s conversion.Scope) error {
	return autoConvert_storage_VolumeAttachment_To_v1alpha1_VolumeAttachment(in, out, s)
}

func autoConvert_v1alpha1_VolumeAttachmentList_To_storage_VolumeAttachmentList(in *v1alpha1.VolumeAttachmentList, out *storage.VolumeAttachmentList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]storage.VolumeAttachment, len(*in))
		for i := range *in {
			if err := Convert_v1alpha1_VolumeAttachment_To_storage_VolumeAttachment(&(*in)[i], &(*out)[i], s); err != nil {
				return err
			}
		}
	} else {
		out.Items = nil
	}
	return nil
}

// Convert_v1alpha1_VolumeAttachmentList_To_storage_VolumeAttachmentList is an autogenerated conversion function.
func Convert_v1alpha1_VolumeAttachmentList_To_storage_VolumeAttachmentList(in *v1alpha1.VolumeAttachmentList, out *storage.VolumeAttachmentList, s conversion.Scope) error {
	return autoConvert_v1alpha1_VolumeAttachmentList_To_storage_VolumeAttachmentList(in, out, s)
}

func autoConvert_storage_VolumeAttachmentList_To_v1alpha1_VolumeAttachmentList(in *storage.VolumeAttachmentList, out *v1alpha1.VolumeAttachmentList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]v1alpha1.VolumeAttachment, len(*in))
		for i := range *in {
			if err := Convert_storage_VolumeAttachment_To_v1alpha1_VolumeAttachment(&(*in)[i], &(*out)[i], s); err != nil {
				return err
			}
		}
	} else {
		out.Items = nil
	}
	return nil
}

// Convert_storage_VolumeAttachmentList_To_v1alpha1_VolumeAttachmentList is an autogenerated conversion function.
func Convert_storage_VolumeAttachmentList_To_v1alpha1_VolumeAttachmentList(in *storage.VolumeAttachmentList, out *v1alpha1.VolumeAttachmentList, s conversion.Scope) error {
	return autoConvert_storage_VolumeAttachmentList_To_v1alpha1_VolumeAttachmentList(in, out, s)
}

func autoConvert_v1alpha1_VolumeAttachmentSource_To_storage_VolumeAttachmentSource(in *v1alpha1.VolumeAttachmentSource, out *storage.VolumeAttachmentSource, s conversion.Scope) error {
	out.PersistentVolumeName = (*string)(unsafe.Pointer(in.PersistentVolumeName))
	if in.InlineVolumeSpec != nil {
		in, out := &in.InlineVolumeSpec, &out.InlineVolumeSpec
		*out = new(core.PersistentVolumeSpec)
		if err := v1.Convert_v1_PersistentVolumeSpec_To_core_PersistentVolumeSpec(*in, *out, s); err != nil {
			return err
		}
	} else {
		out.InlineVolumeSpec = nil
	}
	return nil
}

// Convert_v1alpha1_VolumeAttachmentSource_To_storage_VolumeAttachmentSource is an autogenerated conversion function.
func Convert_v1alpha1_VolumeAttachmentSource_To_storage_VolumeAttachmentSource(in *v1alpha1.VolumeAttachmentSource, out *storage.VolumeAttachmentSource, s conversion.Scope) error {
	return autoConvert_v1alpha1_VolumeAttachmentSource_To_storage_VolumeAttachmentSource(in, out, s)
}

func autoConvert_storage_VolumeAttachmentSource_To_v1alpha1_VolumeAttachmentSource(in *storage.VolumeAttachmentSource, out *v1alpha1.VolumeAttachmentSource, s conversion.Scope) error {
	out.PersistentVolumeName = (*string)(unsafe.Pointer(in.PersistentVolumeName))
	if in.InlineVolumeSpec != nil {
		in, out := &in.InlineVolumeSpec, &out.InlineVolumeSpec
		*out = new(corev1.PersistentVolumeSpec)
		if err := v1.Convert_core_PersistentVolumeSpec_To_v1_PersistentVolumeSpec(*in, *out, s); err != nil {
			return err
		}
	} else {
		out.InlineVolumeSpec = nil
	}
	return nil
}

// Convert_storage_VolumeAttachmentSource_To_v1alpha1_VolumeAttachmentSource is an autogenerated conversion function.
func Convert_storage_VolumeAttachmentSource_To_v1alpha1_VolumeAttachmentSource(in *storage.VolumeAttachmentSource, out *v1alpha1.VolumeAttachmentSource, s conversion.Scope) error {
	return autoConvert_storage_VolumeAttachmentSource_To_v1alpha1_VolumeAttachmentSource(in, out, s)
}

func autoConvert_v1alpha1_VolumeAttachmentSpec_To_storage_VolumeAttachmentSpec(in *v1alpha1.VolumeAttachmentSpec, out *storage.VolumeAttachmentSpec, s conversion.Scope) error {
	out.Attacher = in.Attacher
	if err := Convert_v1alpha1_VolumeAttachmentSource_To_storage_VolumeAttachmentSource(&in.Source, &out.Source, s); err != nil {
		return err
	}
	out.NodeName = in.NodeName
	return nil
}

// Convert_v1alpha1_VolumeAttachmentSpec_To_storage_VolumeAttachmentSpec is an autogenerated conversion function.
func Convert_v1alpha1_VolumeAttachmentSpec_To_storage_VolumeAttachmentSpec(in *v1alpha1.VolumeAttachmentSpec, out *storage.VolumeAttachmentSpec, s conversion.Scope) error {
	return autoConvert_v1alpha1_VolumeAttachmentSpec_To_storage_VolumeAttachmentSpec(in, out, s)
}

func autoConvert_storage_VolumeAttachmentSpec_To_v1alpha1_VolumeAttachmentSpec(in *storage.VolumeAttachmentSpec, out *v1alpha1.VolumeAttachmentSpec, s conversion.Scope) error {
	out.Attacher = in.Attacher
	if err := Convert_storage_VolumeAttachmentSource_To_v1alpha1_VolumeAttachmentSource(&in.Source, &out.Source, s); err != nil {
		return err
	}
	out.NodeName = in.NodeName
	return nil
}

// Convert_storage_VolumeAttachmentSpec_To_v1alpha1_VolumeAttachmentSpec is an autogenerated conversion function.
func Convert_storage_VolumeAttachmentSpec_To_v1alpha1_VolumeAttachmentSpec(in *storage.VolumeAttachmentSpec, out *v1alpha1.VolumeAttachmentSpec, s conversion.Scope) error {
	return autoConvert_storage_VolumeAttachmentSpec_To_v1alpha1_VolumeAttachmentSpec(in, out, s)
}

func autoConvert_v1alpha1_VolumeAttachmentStatus_To_storage_VolumeAttachmentStatus(in *v1alpha1.VolumeAttachmentStatus, out *storage.VolumeAttachmentStatus, s conversion.Scope) error {
	out.Attached = in.Attached
	out.AttachmentMetadata = *(*map[string]string)(unsafe.Pointer(&in.AttachmentMetadata))
	out.AttachError = (*storage.VolumeError)(unsafe.Pointer(in.AttachError))
	out.DetachError = (*storage.VolumeError)(unsafe.Pointer(in.DetachError))
	return nil
}

// Convert_v1alpha1_VolumeAttachmentStatus_To_storage_VolumeAttachmentStatus is an autogenerated conversion function.
func Convert_v1alpha1_VolumeAttachmentStatus_To_storage_VolumeAttachmentStatus(in *v1alpha1.VolumeAttachmentStatus, out *storage.VolumeAttachmentStatus, s conversion.Scope) error {
	return autoConvert_v1alpha1_VolumeAttachmentStatus_To_storage_VolumeAttachmentStatus(in, out, s)
}

func autoConvert_storage_VolumeAttachmentStatus_To_v1alpha1_VolumeAttachmentStatus(in *storage.VolumeAttachmentStatus, out *v1alpha1.VolumeAttachmentStatus, s conversion.Scope) error {
	out.Attached = in.Attached
	out.AttachmentMetadata = *(*map[string]string)(unsafe.Pointer(&in.AttachmentMetadata))
	out.AttachError = (*v1alpha1.VolumeError)(unsafe.Pointer(in.AttachError))
	out.DetachError = (*v1alpha1.VolumeError)(unsafe.Pointer(in.DetachError))
	return nil
}

// Convert_storage_VolumeAttachmentStatus_To_v1alpha1_VolumeAttachmentStatus is an autogenerated conversion function.
func Convert_storage_VolumeAttachmentStatus_To_v1alpha1_VolumeAttachmentStatus(in *storage.VolumeAttachmentStatus, out *v1alpha1.VolumeAttachmentStatus, s conversion.Scope) error {
	return autoConvert_storage_VolumeAttachmentStatus_To_v1alpha1_VolumeAttachmentStatus(in, out, s)
}

func autoConvert_v1alpha1_VolumeError_To_storage_VolumeError(in *v1alpha1.VolumeError, out *storage.VolumeError, s conversion.Scope) error {
	out.Time = in.Time
	out.Message = in.Message
	return nil
}

// Convert_v1alpha1_VolumeError_To_storage_VolumeError is an autogenerated conversion function.
func Convert_v1alpha1_VolumeError_To_storage_VolumeError(in *v1alpha1.VolumeError, out *storage.VolumeError, s conversion.Scope) error {
	return autoConvert_v1alpha1_VolumeError_To_storage_VolumeError(in, out, s)
}

func autoConvert_storage_VolumeError_To_v1alpha1_VolumeError(in *storage.VolumeError, out *v1alpha1.VolumeError, s conversion.Scope) error {
	out.Time = in.Time
	out.Message = in.Message
	return nil
}

// Convert_storage_VolumeError_To_v1alpha1_VolumeError is an autogenerated conversion function.
func Convert_storage_VolumeError_To_v1alpha1_VolumeError(in *storage.VolumeError, out *v1alpha1.VolumeError, s conversion.Scope) error {
	return autoConvert_storage_VolumeError_To_v1alpha1_VolumeError(in, out, s)
}
