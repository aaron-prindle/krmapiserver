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

package rest

import (
	"fmt"
	"time"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/klog"

	schedulingv1beta1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/scheduling/v1beta1"
	apierrors "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/api/errors"
	metav1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/wait"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/generic"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/server"
	serverstorage "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/server/storage"
	schedulingclient "github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/kubernetes/typed/scheduling/v1beta1"
	"github.com/aaron-prindle/krmapiserver/pkg/api/legacyscheme"
	"github.com/aaron-prindle/krmapiserver/pkg/apis/scheduling"
	schedulingapiv1 "github.com/aaron-prindle/krmapiserver/pkg/apis/scheduling/v1"
	schedulingapiv1alpha1 "github.com/aaron-prindle/krmapiserver/pkg/apis/scheduling/v1alpha1"
	schedulingapiv1beta1 "github.com/aaron-prindle/krmapiserver/pkg/apis/scheduling/v1beta1"
	priorityclassstore "github.com/aaron-prindle/krmapiserver/pkg/registry/scheduling/priorityclass/storage"
)

const PostStartHookName = "scheduling/bootstrap-system-priority-classes"

type RESTStorageProvider struct{}

var _ genericapiserver.PostStartHookProvider = RESTStorageProvider{}

func (p RESTStorageProvider) NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (genericapiserver.APIGroupInfo, bool) {
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(scheduling.GroupName, legacyscheme.Scheme, legacyscheme.ParameterCodec, legacyscheme.Codecs)

	if apiResourceConfigSource.VersionEnabled(schedulingapiv1alpha1.SchemeGroupVersion) {
		apiGroupInfo.VersionedResourcesStorageMap[schedulingapiv1alpha1.SchemeGroupVersion.Version] = p.v1alpha1Storage(apiResourceConfigSource, restOptionsGetter)
	}
	if apiResourceConfigSource.VersionEnabled(schedulingapiv1beta1.SchemeGroupVersion) {
		apiGroupInfo.VersionedResourcesStorageMap[schedulingapiv1beta1.SchemeGroupVersion.Version] = p.v1beta1Storage(apiResourceConfigSource, restOptionsGetter)
	}
	if apiResourceConfigSource.VersionEnabled(schedulingapiv1.SchemeGroupVersion) {
		apiGroupInfo.VersionedResourcesStorageMap[schedulingapiv1.SchemeGroupVersion.Version] = p.v1Storage(apiResourceConfigSource, restOptionsGetter)
	}
	return apiGroupInfo, true
}

func (p RESTStorageProvider) v1alpha1Storage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) map[string]rest.Storage {
	storage := map[string]rest.Storage{}
	// priorityclasses
	priorityClassStorage := priorityclassstore.NewREST(restOptionsGetter)
	storage["priorityclasses"] = priorityClassStorage

	return storage
}

func (p RESTStorageProvider) v1beta1Storage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) map[string]rest.Storage {
	storage := map[string]rest.Storage{}
	// priorityclasses
	priorityClassStorage := priorityclassstore.NewREST(restOptionsGetter)
	storage["priorityclasses"] = priorityClassStorage

	return storage
}

func (p RESTStorageProvider) v1Storage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) map[string]rest.Storage {
	storage := map[string]rest.Storage{}
	// priorityclasses
	priorityClassStorage := priorityclassstore.NewREST(restOptionsGetter)
	storage["priorityclasses"] = priorityClassStorage

	return storage
}

func (p RESTStorageProvider) PostStartHook() (string, genericapiserver.PostStartHookFunc, error) {
	return PostStartHookName, AddSystemPriorityClasses(), nil
}

func AddSystemPriorityClasses() genericapiserver.PostStartHookFunc {
	return func(hookContext genericapiserver.PostStartHookContext) error {
		// Adding system priority classes is important. If they fail to add, many critical system
		// components may fail and cluster may break.
		err := wait.Poll(1*time.Second, 30*time.Second, func() (done bool, err error) {
			schedClientSet, err := schedulingclient.NewForConfig(hookContext.LoopbackClientConfig)
			if err != nil {
				utilruntime.HandleError(fmt.Errorf("unable to initialize client: %v", err))
				return false, nil
			}

			for _, pc := range scheduling.SystemPriorityClasses() {
				_, err := schedClientSet.PriorityClasses().Get(pc.Name, metav1.GetOptions{})
				if err != nil {
					if apierrors.IsNotFound(err) {
						// TODO: Remove this explicit conversion after scheduling api move to v1
						v1beta1PriorityClass := &schedulingv1beta1.PriorityClass{}
						if err := schedulingapiv1beta1.Convert_scheduling_PriorityClass_To_v1beta1_PriorityClass(pc, v1beta1PriorityClass, nil); err != nil {
							return false, err
						}
						_, err := schedClientSet.PriorityClasses().Create(v1beta1PriorityClass)
						if err != nil && !apierrors.IsAlreadyExists(err) {
							return false, err
						} else {
							klog.Infof("created PriorityClass %s with value %v", pc.Name, pc.Value)
						}
					} else {
						// Unable to get the priority class for reasons other than "not found".
						klog.Warningf("unable to get PriorityClass %v: %v. Retrying...", pc.Name, err)
						return false, nil
					}
				}
			}
			klog.Infof("all system priority classes are created successfully or already exist.")
			return true, nil
		})
		// if we're never able to make it through initialization, kill the API server.
		if err != nil {
			return fmt.Errorf("unable to add default system priority classes: %v", err)
		}
		return nil
	}
}

func (p RESTStorageProvider) GroupName() string {
	return scheduling.GroupName
}
