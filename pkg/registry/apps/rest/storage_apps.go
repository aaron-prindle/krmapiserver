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
	appsapiv1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/apps/v1"
	appsapiv1beta1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/apps/v1beta1"
	appsapiv1beta2 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/apps/v1beta2"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/generic"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/server"
	serverstorage "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/server/storage"
	"github.com/aaron-prindle/krmapiserver/pkg/api/legacyscheme"
	"github.com/aaron-prindle/krmapiserver/pkg/apis/apps"
	controllerrevisionsstore "github.com/aaron-prindle/krmapiserver/pkg/registry/apps/controllerrevision/storage"
	daemonsetstore "github.com/aaron-prindle/krmapiserver/pkg/registry/apps/daemonset/storage"
	deploymentstore "github.com/aaron-prindle/krmapiserver/pkg/registry/apps/deployment/storage"
	replicasetstore "github.com/aaron-prindle/krmapiserver/pkg/registry/apps/replicaset/storage"
	statefulsetstore "github.com/aaron-prindle/krmapiserver/pkg/registry/apps/statefulset/storage"
)

type RESTStorageProvider struct{}

func (p RESTStorageProvider) NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (genericapiserver.APIGroupInfo, bool) {
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(apps.GroupName, legacyscheme.Scheme, legacyscheme.ParameterCodec, legacyscheme.Codecs)
	// If you add a version here, be sure to add an entry in `k8s.io/kubernetes/cmd/kube-apiserver/app/aggregator.go with specific priorities.
	// TODO refactor the plumbing to provide the information in the APIGroupInfo

	if apiResourceConfigSource.VersionEnabled(appsapiv1beta1.SchemeGroupVersion) {
		apiGroupInfo.VersionedResourcesStorageMap[appsapiv1beta1.SchemeGroupVersion.Version] = p.v1beta1Storage(apiResourceConfigSource, restOptionsGetter)
	}
	if apiResourceConfigSource.VersionEnabled(appsapiv1beta2.SchemeGroupVersion) {
		apiGroupInfo.VersionedResourcesStorageMap[appsapiv1beta2.SchemeGroupVersion.Version] = p.v1beta2Storage(apiResourceConfigSource, restOptionsGetter)
	}
	if apiResourceConfigSource.VersionEnabled(appsapiv1.SchemeGroupVersion) {
		apiGroupInfo.VersionedResourcesStorageMap[appsapiv1.SchemeGroupVersion.Version] = p.v1Storage(apiResourceConfigSource, restOptionsGetter)
	}

	return apiGroupInfo, true
}

func (p RESTStorageProvider) v1beta1Storage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) map[string]rest.Storage {
	storage := map[string]rest.Storage{}

	// deployments
	deploymentStorage := deploymentstore.NewStorage(restOptionsGetter)
	storage["deployments"] = deploymentStorage.Deployment
	storage["deployments/status"] = deploymentStorage.Status
	storage["deployments/rollback"] = deploymentStorage.Rollback
	storage["deployments/scale"] = deploymentStorage.Scale

	// statefulsets
	statefulSetStorage := statefulsetstore.NewStorage(restOptionsGetter)
	storage["statefulsets"] = statefulSetStorage.StatefulSet
	storage["statefulsets/status"] = statefulSetStorage.Status
	storage["statefulsets/scale"] = statefulSetStorage.Scale

	// controllerrevisions
	historyStorage := controllerrevisionsstore.NewREST(restOptionsGetter)
	storage["controllerrevisions"] = historyStorage

	return storage
}

func (p RESTStorageProvider) v1beta2Storage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) map[string]rest.Storage {
	storage := map[string]rest.Storage{}

	// deployments
	deploymentStorage := deploymentstore.NewStorage(restOptionsGetter)
	storage["deployments"] = deploymentStorage.Deployment
	storage["deployments/status"] = deploymentStorage.Status
	storage["deployments/scale"] = deploymentStorage.Scale

	// statefulsets
	statefulSetStorage := statefulsetstore.NewStorage(restOptionsGetter)
	storage["statefulsets"] = statefulSetStorage.StatefulSet
	storage["statefulsets/status"] = statefulSetStorage.Status
	storage["statefulsets/scale"] = statefulSetStorage.Scale

	// daemonsets
	daemonSetStorage, daemonSetStatusStorage := daemonsetstore.NewREST(restOptionsGetter)
	storage["daemonsets"] = daemonSetStorage
	storage["daemonsets/status"] = daemonSetStatusStorage

	// replicasets
	replicaSetStorage := replicasetstore.NewStorage(restOptionsGetter)
	storage["replicasets"] = replicaSetStorage.ReplicaSet
	storage["replicasets/status"] = replicaSetStorage.Status
	storage["replicasets/scale"] = replicaSetStorage.Scale

	// controllerrevisions
	historyStorage := controllerrevisionsstore.NewREST(restOptionsGetter)
	storage["controllerrevisions"] = historyStorage

	return storage
}

func (p RESTStorageProvider) v1Storage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) map[string]rest.Storage {
	storage := map[string]rest.Storage{}

	// deployments
	deploymentStorage := deploymentstore.NewStorage(restOptionsGetter)
	storage["deployments"] = deploymentStorage.Deployment
	storage["deployments/status"] = deploymentStorage.Status
	storage["deployments/scale"] = deploymentStorage.Scale

	// statefulsets
	statefulSetStorage := statefulsetstore.NewStorage(restOptionsGetter)
	storage["statefulsets"] = statefulSetStorage.StatefulSet
	storage["statefulsets/status"] = statefulSetStorage.Status
	storage["statefulsets/scale"] = statefulSetStorage.Scale

	// daemonsets
	daemonSetStorage, daemonSetStatusStorage := daemonsetstore.NewREST(restOptionsGetter)
	storage["daemonsets"] = daemonSetStorage
	storage["daemonsets/status"] = daemonSetStatusStorage

	// replicasets
	replicaSetStorage := replicasetstore.NewStorage(restOptionsGetter)
	storage["replicasets"] = replicaSetStorage.ReplicaSet
	storage["replicasets/status"] = replicaSetStorage.Status
	storage["replicasets/scale"] = replicaSetStorage.Scale

	// controllerrevisions
	historyStorage := controllerrevisionsstore.NewREST(restOptionsGetter)
	storage["controllerrevisions"] = historyStorage

	return storage
}

func (p RESTStorageProvider) GroupName() string {
	return apps.GroupName
}
