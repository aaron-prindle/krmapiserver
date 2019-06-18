/*
Copyright 2018 The Kubernetes Authors.

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

package admission

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/klog"

	utilwait "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/wait"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/admission"
	webhookinit "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/admission/plugin/webhook/initializer"
	genericapiserver "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/server"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/util/webhook"
	cacheddiscovery "github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/discovery/cached/memory"
	externalinformers "github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/informers"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/kubernetes"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/rest"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/restmapper"
	quotainstall "github.com/aaron-prindle/krmapiserver/pkg/quota/v1/install"
)

// Config holds the configuration needed to for initialize the admission plugins
type Config struct {
	CloudConfigFile      string
	LoopbackClientConfig *rest.Config
	ExternalInformers    externalinformers.SharedInformerFactory
}

// New sets up the plugins and admission start hooks needed for admission
func (c *Config) New(proxyTransport *http.Transport, serviceResolver webhook.ServiceResolver) ([]admission.PluginInitializer, genericapiserver.PostStartHookFunc, error) {
	webhookAuthResolverWrapper := webhook.NewDefaultAuthenticationInfoResolverWrapper(proxyTransport, c.LoopbackClientConfig)
	webhookPluginInitializer := webhookinit.NewPluginInitializer(webhookAuthResolverWrapper, serviceResolver)

	var cloudConfig []byte
	if c.CloudConfigFile != "" {
		var err error
		cloudConfig, err = ioutil.ReadFile(c.CloudConfigFile)
		if err != nil {
			klog.Fatalf("Error reading from cloud configuration file %s: %#v", c.CloudConfigFile, err)
		}
	}
	clientset, err := kubernetes.NewForConfig(c.LoopbackClientConfig)
	if err != nil {
		return nil, nil, err
	}

	discoveryClient := cacheddiscovery.NewMemCacheClient(clientset.Discovery())
	discoveryRESTMapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	kubePluginInitializer := NewPluginInitializer(
		cloudConfig,
		discoveryRESTMapper,
		quotainstall.NewQuotaConfigurationForAdmission(),
	)

	admissionPostStartHook := func(context genericapiserver.PostStartHookContext) error {
		discoveryRESTMapper.Reset()
		go utilwait.Until(discoveryRESTMapper.Reset, 30*time.Second, context.StopCh)
		return nil
	}

	return []admission.PluginInitializer{webhookPluginInitializer, kubePluginInitializer}, admissionPostStartHook, nil
}
