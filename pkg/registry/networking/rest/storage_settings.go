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
	networkingapiv1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/networking/v1"
	networkingapiv1beta1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/networking/v1beta1"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/generic"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/server"
	serverstorage "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/server/storage"
	"github.com/aaron-prindle/krmapiserver/pkg/api/legacyscheme"
	"github.com/aaron-prindle/krmapiserver/pkg/apis/networking"
	ingressstore "github.com/aaron-prindle/krmapiserver/pkg/registry/networking/ingress/storage"
	networkpolicystore "github.com/aaron-prindle/krmapiserver/pkg/registry/networking/networkpolicy/storage"
)

type RESTStorageProvider struct{}

func (p RESTStorageProvider) NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (genericapiserver.APIGroupInfo, bool) {
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(networking.GroupName, legacyscheme.Scheme, legacyscheme.ParameterCodec, legacyscheme.Codecs)
	// If you add a version here, be sure to add an entry in `k8s.io/kubernetes/cmd/kube-apiserver/app/aggregator.go with specific priorities.
	// TODO refactor the plumbing to provide the information in the APIGroupInfo

	if apiResourceConfigSource.VersionEnabled(networkingapiv1.SchemeGroupVersion) {
		apiGroupInfo.VersionedResourcesStorageMap[networkingapiv1.SchemeGroupVersion.Version] = p.v1Storage(apiResourceConfigSource, restOptionsGetter)
	}

	if apiResourceConfigSource.VersionEnabled(networkingapiv1beta1.SchemeGroupVersion) {
		apiGroupInfo.VersionedResourcesStorageMap[networkingapiv1beta1.SchemeGroupVersion.Version] = p.v1beta1Storage(apiResourceConfigSource, restOptionsGetter)
	}

	return apiGroupInfo, true
}

func (p RESTStorageProvider) v1Storage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) map[string]rest.Storage {
	storage := map[string]rest.Storage{}
	// networkpolicies
	networkPolicyStorage := networkpolicystore.NewREST(restOptionsGetter)
	storage["networkpolicies"] = networkPolicyStorage

	return storage
}

func (p RESTStorageProvider) v1beta1Storage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) map[string]rest.Storage {
	storage := map[string]rest.Storage{}
	// ingresses
	ingressStorage, ingressStatusStorage := ingressstore.NewREST(restOptionsGetter)
	storage["ingresses"] = ingressStorage
	storage["ingresses/status"] = ingressStatusStorage
	return storage
}

func (p RESTStorageProvider) GroupName() string {
	return networking.GroupName
}
