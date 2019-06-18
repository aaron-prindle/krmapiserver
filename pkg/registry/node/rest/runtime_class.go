/*
Copyright 2019 The Kubernetes Authors.

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
	nodev1alpha1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/node/v1alpha1"
	nodev1beta1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/node/v1beta1"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/generic"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/server"
	serverstorage "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/server/storage"
	"github.com/aaron-prindle/krmapiserver/pkg/api/legacyscheme"
	nodeinternal "github.com/aaron-prindle/krmapiserver/pkg/apis/node"
	runtimeclassstorage "github.com/aaron-prindle/krmapiserver/pkg/registry/node/runtimeclass/storage"
)

// RESTStorageProvider is a REST storage provider for node.k8s.io
type RESTStorageProvider struct{}

// NewRESTStorage returns a RESTStorageProvider
func (p RESTStorageProvider) NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (genericapiserver.APIGroupInfo, bool) {
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(nodeinternal.GroupName, legacyscheme.Scheme, legacyscheme.ParameterCodec, legacyscheme.Codecs)

	if apiResourceConfigSource.VersionEnabled(nodev1alpha1.SchemeGroupVersion) {
		apiGroupInfo.VersionedResourcesStorageMap[nodev1alpha1.SchemeGroupVersion.Version] = p.v1alpha1Storage(apiResourceConfigSource, restOptionsGetter)
	}

	if apiResourceConfigSource.VersionEnabled(nodev1beta1.SchemeGroupVersion) {
		apiGroupInfo.VersionedResourcesStorageMap[nodev1beta1.SchemeGroupVersion.Version] = p.v1beta1Storage(apiResourceConfigSource, restOptionsGetter)
	}

	return apiGroupInfo, true
}

func (p RESTStorageProvider) v1alpha1Storage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) map[string]rest.Storage {
	storage := map[string]rest.Storage{}
	s := runtimeclassstorage.NewREST(restOptionsGetter)
	storage["runtimeclasses"] = s

	return storage
}

func (p RESTStorageProvider) v1beta1Storage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) map[string]rest.Storage {
	storage := map[string]rest.Storage{}
	s := runtimeclassstorage.NewREST(restOptionsGetter)
	storage["runtimeclasses"] = s

	return storage
}

// GroupName is the group name for the storage provider
func (p RESTStorageProvider) GroupName() string {
	return nodeinternal.GroupName
}
