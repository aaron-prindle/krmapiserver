/*
Copyright 2017 The Kubernetes Authors.

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

package v1beta1

import (
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/api/core/v1"
	storagev1beta1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/storage/v1beta1"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

func SetDefaults_StorageClass(obj *storagev1beta1.StorageClass) {
	if obj.ReclaimPolicy == nil {
		obj.ReclaimPolicy = new(v1.PersistentVolumeReclaimPolicy)
		*obj.ReclaimPolicy = v1.PersistentVolumeReclaimDelete
	}

	if obj.VolumeBindingMode == nil {
		obj.VolumeBindingMode = new(storagev1beta1.VolumeBindingMode)
		*obj.VolumeBindingMode = storagev1beta1.VolumeBindingImmediate
	}
}

func SetDefaults_CSIDriver(obj *storagev1beta1.CSIDriver) {
	if obj.Spec.AttachRequired == nil {
		obj.Spec.AttachRequired = new(bool)
		*(obj.Spec.AttachRequired) = true
	}
	if obj.Spec.PodInfoOnMount == nil {
		obj.Spec.PodInfoOnMount = new(bool)
		*(obj.Spec.PodInfoOnMount) = false
	}
}
