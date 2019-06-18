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

// Package app does all of the work necessary to create a Kubernetes
// APIServer by binding together the API, master and APIServer infrastructure.
// It can be configured and called directly or via the hyperkube framework.
package app

import (
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsapiserver "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiextensions-apiserver/pkg/apiserver"
	apiextensionsoptions "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiextensions-apiserver/pkg/cmd/server/options"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime/schema"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/admission"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/features"
	genericapiserver "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/server"
	genericoptions "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/server/options"
	utilfeature "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/util/feature"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/util/webhook"
	kubeexternalinformers "github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/informers"
	"github.com/aaron-prindle/krmapiserver/cmd/kube-apiserver/app/options"
)

func createAPIExtensionsConfig(
	kubeAPIServerConfig genericapiserver.Config,
	externalInformers kubeexternalinformers.SharedInformerFactory,
	pluginInitializers []admission.PluginInitializer,
	commandOptions *options.ServerRunOptions,
	masterCount int,
	serviceResolver webhook.ServiceResolver,
	authResolverWrapper webhook.AuthenticationInfoResolverWrapper,
) (*apiextensionsapiserver.Config, error) {
	// make a shallow copy to let us twiddle a few things
	// most of the config actually remains the same.  We only need to mess with a couple items related to the particulars of the apiextensions
	genericConfig := kubeAPIServerConfig

	// override genericConfig.AdmissionControl with apiextensions' scheme,
	// because apiextentions apiserver should use its own scheme to convert resources.
	err := commandOptions.Admission.ApplyTo(
		&genericConfig,
		externalInformers,
		genericConfig.LoopbackClientConfig,
		pluginInitializers...)
	if err != nil {
		return nil, err
	}

	// copy the etcd options so we don't mutate originals.
	etcdOptions := *commandOptions.Etcd
	etcdOptions.StorageConfig.Paging = utilfeature.DefaultFeatureGate.Enabled(features.APIListChunking)
	etcdOptions.StorageConfig.Codec = apiextensionsapiserver.Codecs.LegacyCodec(v1beta1.SchemeGroupVersion)
	etcdOptions.StorageConfig.EncodeVersioner = runtime.NewMultiGroupVersioner(v1beta1.SchemeGroupVersion, schema.GroupKind{Group: v1beta1.GroupName})
	genericConfig.RESTOptionsGetter = &genericoptions.SimpleRestOptionsFactory{Options: etcdOptions}

	// override MergedResourceConfig with apiextensions defaults and registry
	if err := commandOptions.APIEnablement.ApplyTo(
		&genericConfig,
		apiextensionsapiserver.DefaultAPIResourceConfigSource(),
		apiextensionsapiserver.Scheme); err != nil {
		return nil, err
	}

	apiextensionsConfig := &apiextensionsapiserver.Config{
		GenericConfig: &genericapiserver.RecommendedConfig{
			Config:                genericConfig,
			SharedInformerFactory: externalInformers,
		},
		ExtraConfig: apiextensionsapiserver.ExtraConfig{
			CRDRESTOptionsGetter: apiextensionsoptions.NewCRDRESTOptionsGetter(etcdOptions),
			MasterCount:          masterCount,
			AuthResolverWrapper:  authResolverWrapper,
			ServiceResolver:      serviceResolver,
		},
	}

	return apiextensionsConfig, nil
}

func createAPIExtensionsServer(apiextensionsConfig *apiextensionsapiserver.Config, delegateAPIServer genericapiserver.DelegationTarget) (*apiextensionsapiserver.CustomResourceDefinitions, error) {
	return apiextensionsConfig.Complete().New(delegateAPIServer)
}
