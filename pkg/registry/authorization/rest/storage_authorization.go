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
	authorizationv1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/authorization/v1"
	authorizationv1beta1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/authorization/v1beta1"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/authorization/authorizer"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/generic"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/server"
	serverstorage "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/server/storage"
	"github.com/aaron-prindle/krmapiserver/pkg/api/legacyscheme"
	"github.com/aaron-prindle/krmapiserver/pkg/apis/authorization"
	"github.com/aaron-prindle/krmapiserver/pkg/registry/authorization/localsubjectaccessreview"
	"github.com/aaron-prindle/krmapiserver/pkg/registry/authorization/selfsubjectaccessreview"
	"github.com/aaron-prindle/krmapiserver/pkg/registry/authorization/selfsubjectrulesreview"
	"github.com/aaron-prindle/krmapiserver/pkg/registry/authorization/subjectaccessreview"
)

type RESTStorageProvider struct {
	Authorizer   authorizer.Authorizer
	RuleResolver authorizer.RuleResolver
}

func (p RESTStorageProvider) NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (genericapiserver.APIGroupInfo, bool) {
	if p.Authorizer == nil {
		return genericapiserver.APIGroupInfo{}, false
	}

	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(authorization.GroupName, legacyscheme.Scheme, legacyscheme.ParameterCodec, legacyscheme.Codecs)
	// If you add a version here, be sure to add an entry in `k8s.io/kubernetes/cmd/kube-apiserver/app/aggregator.go with specific priorities.
	// TODO refactor the plumbing to provide the information in the APIGroupInfo

	if apiResourceConfigSource.VersionEnabled(authorizationv1beta1.SchemeGroupVersion) {
		apiGroupInfo.VersionedResourcesStorageMap[authorizationv1beta1.SchemeGroupVersion.Version] = p.v1beta1Storage(apiResourceConfigSource, restOptionsGetter)
	}

	if apiResourceConfigSource.VersionEnabled(authorizationv1.SchemeGroupVersion) {
		apiGroupInfo.VersionedResourcesStorageMap[authorizationv1.SchemeGroupVersion.Version] = p.v1Storage(apiResourceConfigSource, restOptionsGetter)
	}

	return apiGroupInfo, true
}

func (p RESTStorageProvider) v1beta1Storage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) map[string]rest.Storage {
	storage := map[string]rest.Storage{}
	// subjectaccessreviews
	storage["subjectaccessreviews"] = subjectaccessreview.NewREST(p.Authorizer)
	// selfsubjectaccessreviews
	storage["selfsubjectaccessreviews"] = selfsubjectaccessreview.NewREST(p.Authorizer)
	// localsubjectaccessreviews
	storage["localsubjectaccessreviews"] = localsubjectaccessreview.NewREST(p.Authorizer)
	// selfsubjectrulesreviews
	storage["selfsubjectrulesreviews"] = selfsubjectrulesreview.NewREST(p.RuleResolver)

	return storage
}

func (p RESTStorageProvider) v1Storage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) map[string]rest.Storage {
	storage := map[string]rest.Storage{}
	// subjectaccessreviews
	storage["subjectaccessreviews"] = subjectaccessreview.NewREST(p.Authorizer)
	// selfsubjectaccessreviews
	storage["selfsubjectaccessreviews"] = selfsubjectaccessreview.NewREST(p.Authorizer)
	// localsubjectaccessreviews
	storage["localsubjectaccessreviews"] = localsubjectaccessreview.NewREST(p.Authorizer)
	// selfsubjectrulesreviews
	storage["selfsubjectrulesreviews"] = selfsubjectrulesreview.NewREST(p.RuleResolver)

	return storage
}

func (p RESTStorageProvider) GroupName() string {
	return authorization.GroupName
}
