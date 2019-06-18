/*
Copyright 2016 The Kubernetes Authors.

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

package rest

import (
	storageapiv1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/storage/v1"
	storageapiv1alpha1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/storage/v1alpha1"
	storageapiv1beta1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/storage/v1beta1"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/generic"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/server"
	serverstorage "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/server/storage"
	utilfeature "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/util/feature"
	"github.com/aaron-prindle/krmapiserver/pkg/api/legacyscheme"
	storageapi "github.com/aaron-prindle/krmapiserver/pkg/apis/storage"
	"github.com/aaron-prindle/krmapiserver/pkg/features"
	csidriverstore "github.com/aaron-prindle/krmapiserver/pkg/registry/storage/csidriver/storage"
	csinodestore "github.com/aaron-prindle/krmapiserver/pkg/registry/storage/csinode/storage"
	storageclassstore "github.com/aaron-prindle/krmapiserver/pkg/registry/storage/storageclass/storage"
	volumeattachmentstore "github.com/aaron-prindle/krmapiserver/pkg/registry/storage/volumeattachment/storage"
)

type RESTStorageProvider struct {
}

func (p RESTStorageProvider) NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (genericapiserver.APIGroupInfo, bool) {
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(storageapi.GroupName, legacyscheme.Scheme, legacyscheme.ParameterCodec, legacyscheme.Codecs)
	// If you add a version here, be sure to add an entry in `k8s.io/kubernetes/cmd/kube-apiserver/app/aggregator.go with specific priorities.
	// TODO refactor the plumbing to provide the information in the APIGroupInfo

	if apiResourceConfigSource.VersionEnabled(storageapiv1alpha1.SchemeGroupVersion) {
		apiGroupInfo.VersionedResourcesStorageMap[storageapiv1alpha1.SchemeGroupVersion.Version] = p.v1alpha1Storage(apiResourceConfigSource, restOptionsGetter)
	}
	if apiResourceConfigSource.VersionEnabled(storageapiv1beta1.SchemeGroupVersion) {
		apiGroupInfo.VersionedResourcesStorageMap[storageapiv1beta1.SchemeGroupVersion.Version] = p.v1beta1Storage(apiResourceConfigSource, restOptionsGetter)
	}
	if apiResourceConfigSource.VersionEnabled(storageapiv1.SchemeGroupVersion) {
		apiGroupInfo.VersionedResourcesStorageMap[storageapiv1.SchemeGroupVersion.Version] = p.v1Storage(apiResourceConfigSource, restOptionsGetter)
	}

	return apiGroupInfo, true
}

func (p RESTStorageProvider) v1alpha1Storage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) map[string]rest.Storage {
	storage := map[string]rest.Storage{}
	// volumeattachments
	volumeAttachmentStorage := volumeattachmentstore.NewStorage(restOptionsGetter)
	storage["volumeattachments"] = volumeAttachmentStorage.VolumeAttachment

	return storage
}

func (p RESTStorageProvider) v1beta1Storage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) map[string]rest.Storage {
	storage := map[string]rest.Storage{}
	// storageclasses
	storageClassStorage := storageclassstore.NewREST(restOptionsGetter)
	storage["storageclasses"] = storageClassStorage

	// volumeattachments
	volumeAttachmentStorage := volumeattachmentstore.NewStorage(restOptionsGetter)
	storage["volumeattachments"] = volumeAttachmentStorage.VolumeAttachment

	// register csinodes if CSINodeInfo feature gate is enabled
	if utilfeature.DefaultFeatureGate.Enabled(features.CSINodeInfo) {
		csiNodeStorage := csinodestore.NewStorage(restOptionsGetter)
		storage["csinodes"] = csiNodeStorage.CSINode
	}

	// register csidrivers if CSIDriverRegistry feature gate is enabled
	if utilfeature.DefaultFeatureGate.Enabled(features.CSIDriverRegistry) {
		csiDriverStorage := csidriverstore.NewStorage(restOptionsGetter)
		storage["csidrivers"] = csiDriverStorage.CSIDriver
	}

	return storage
}

func (p RESTStorageProvider) v1Storage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) map[string]rest.Storage {
	storageClassStorage := storageclassstore.NewREST(restOptionsGetter)
	volumeAttachmentStorage := volumeattachmentstore.NewStorage(restOptionsGetter)

	storage := map[string]rest.Storage{
		// storageclasses
		"storageclasses": storageClassStorage,

		// volumeattachments
		"volumeattachments":        volumeAttachmentStorage.VolumeAttachment,
		"volumeattachments/status": volumeAttachmentStorage.Status,
	}

	return storage
}

func (p RESTStorageProvider) GroupName() string {
	return storageapi.GroupName
}
