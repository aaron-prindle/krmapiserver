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
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/klog"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime/schema"
	utilnet "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/net"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/authentication/authenticator"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/generic"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/server"
	serverstorage "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/server/storage"
	etcdutil "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/storage/etcd/util"
	utilfeature "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/util/feature"
	policyclient "github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/kubernetes/typed/policy/v1beta1"
	restclient "github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/rest"
	"github.com/aaron-prindle/krmapiserver/pkg/api/legacyscheme"
	api "github.com/aaron-prindle/krmapiserver/pkg/apis/core"
	"github.com/aaron-prindle/krmapiserver/pkg/features"
	kubeletclient "github.com/aaron-prindle/krmapiserver/pkg/kubelet/client"
	"github.com/aaron-prindle/krmapiserver/pkg/master/ports"
	"github.com/aaron-prindle/krmapiserver/pkg/registry/core/componentstatus"
	configmapstore "github.com/aaron-prindle/krmapiserver/pkg/registry/core/configmap/storage"
	endpointsstore "github.com/aaron-prindle/krmapiserver/pkg/registry/core/endpoint/storage"
	eventstore "github.com/aaron-prindle/krmapiserver/pkg/registry/core/event/storage"
	limitrangestore "github.com/aaron-prindle/krmapiserver/pkg/registry/core/limitrange/storage"
	namespacestore "github.com/aaron-prindle/krmapiserver/pkg/registry/core/namespace/storage"
	nodestore "github.com/aaron-prindle/krmapiserver/pkg/registry/core/node/storage"
	pvstore "github.com/aaron-prindle/krmapiserver/pkg/registry/core/persistentvolume/storage"
	pvcstore "github.com/aaron-prindle/krmapiserver/pkg/registry/core/persistentvolumeclaim/storage"
	podstore "github.com/aaron-prindle/krmapiserver/pkg/registry/core/pod/storage"
	podtemplatestore "github.com/aaron-prindle/krmapiserver/pkg/registry/core/podtemplate/storage"
	"github.com/aaron-prindle/krmapiserver/pkg/registry/core/rangeallocation"
	controllerstore "github.com/aaron-prindle/krmapiserver/pkg/registry/core/replicationcontroller/storage"
	resourcequotastore "github.com/aaron-prindle/krmapiserver/pkg/registry/core/resourcequota/storage"
	secretstore "github.com/aaron-prindle/krmapiserver/pkg/registry/core/secret/storage"
	"github.com/aaron-prindle/krmapiserver/pkg/registry/core/service/allocator"
	serviceallocator "github.com/aaron-prindle/krmapiserver/pkg/registry/core/service/allocator/storage"
	"github.com/aaron-prindle/krmapiserver/pkg/registry/core/service/ipallocator"
	"github.com/aaron-prindle/krmapiserver/pkg/registry/core/service/portallocator"
	servicestore "github.com/aaron-prindle/krmapiserver/pkg/registry/core/service/storage"
	serviceaccountstore "github.com/aaron-prindle/krmapiserver/pkg/registry/core/serviceaccount/storage"
	"github.com/aaron-prindle/krmapiserver/pkg/serviceaccount"
)

// LegacyRESTStorageProvider provides information needed to build RESTStorage for core, but
// does NOT implement the "normal" RESTStorageProvider (yet!)
type LegacyRESTStorageProvider struct {
	StorageFactory serverstorage.StorageFactory
	// Used for custom proxy dialing, and proxy TLS options
	ProxyTransport      http.RoundTripper
	KubeletClientConfig kubeletclient.KubeletClientConfig
	EventTTL            time.Duration

	// ServiceIPRange is used to build cluster IPs for discovery.
	ServiceIPRange       net.IPNet
	ServiceNodePortRange utilnet.PortRange

	ServiceAccountIssuer        serviceaccount.TokenGenerator
	ServiceAccountMaxExpiration time.Duration

	APIAudiences authenticator.Audiences

	LoopbackClientConfig *restclient.Config
}

// LegacyRESTStorage returns stateful information about particular instances of REST storage to
// master.go for wiring controllers.
// TODO remove this by running the controller as a poststarthook
type LegacyRESTStorage struct {
	ServiceClusterIPAllocator rangeallocation.RangeRegistry
	ServiceNodePortAllocator  rangeallocation.RangeRegistry
}

func (c LegacyRESTStorageProvider) NewLegacyRESTStorage(restOptionsGetter generic.RESTOptionsGetter) (LegacyRESTStorage, genericapiserver.APIGroupInfo, error) {
	apiGroupInfo := genericapiserver.APIGroupInfo{
		PrioritizedVersions:          legacyscheme.Scheme.PrioritizedVersionsForGroup(""),
		VersionedResourcesStorageMap: map[string]map[string]rest.Storage{},
		Scheme:                       legacyscheme.Scheme,
		ParameterCodec:               legacyscheme.ParameterCodec,
		NegotiatedSerializer:         legacyscheme.Codecs,
	}

	var podDisruptionClient policyclient.PodDisruptionBudgetsGetter
	if policyGroupVersion := (schema.GroupVersion{Group: "policy", Version: "v1beta1"}); legacyscheme.Scheme.IsVersionRegistered(policyGroupVersion) {
		var err error
		podDisruptionClient, err = policyclient.NewForConfig(c.LoopbackClientConfig)
		if err != nil {
			return LegacyRESTStorage{}, genericapiserver.APIGroupInfo{}, err
		}
	}
	restStorage := LegacyRESTStorage{}

	podTemplateStorage := podtemplatestore.NewREST(restOptionsGetter)

	eventStorage := eventstore.NewREST(restOptionsGetter, uint64(c.EventTTL.Seconds()))
	limitRangeStorage := limitrangestore.NewREST(restOptionsGetter)

	resourceQuotaStorage, resourceQuotaStatusStorage := resourcequotastore.NewREST(restOptionsGetter)
	secretStorage := secretstore.NewREST(restOptionsGetter)
	persistentVolumeStorage, persistentVolumeStatusStorage := pvstore.NewREST(restOptionsGetter)
	persistentVolumeClaimStorage, persistentVolumeClaimStatusStorage := pvcstore.NewREST(restOptionsGetter)
	configMapStorage := configmapstore.NewREST(restOptionsGetter)

	namespaceStorage, namespaceStatusStorage, namespaceFinalizeStorage := namespacestore.NewREST(restOptionsGetter)

	endpointsStorage := endpointsstore.NewREST(restOptionsGetter)

	nodeStorage, err := nodestore.NewStorage(restOptionsGetter, c.KubeletClientConfig, c.ProxyTransport)
	if err != nil {
		return LegacyRESTStorage{}, genericapiserver.APIGroupInfo{}, err
	}

	podStorage := podstore.NewStorage(
		restOptionsGetter,
		nodeStorage.KubeletConnectionInfo,
		c.ProxyTransport,
		podDisruptionClient,
	)

	var serviceAccountStorage *serviceaccountstore.REST
	if c.ServiceAccountIssuer != nil && utilfeature.DefaultFeatureGate.Enabled(features.TokenRequest) {
		serviceAccountStorage = serviceaccountstore.NewREST(restOptionsGetter, c.ServiceAccountIssuer, c.APIAudiences, c.ServiceAccountMaxExpiration, podStorage.Pod.Store, secretStorage.Store)
	} else {
		serviceAccountStorage = serviceaccountstore.NewREST(restOptionsGetter, nil, nil, 0, nil, nil)
	}

	serviceRESTStorage, serviceStatusStorage := servicestore.NewGenericREST(restOptionsGetter)

	var serviceClusterIPRegistry rangeallocation.RangeRegistry
	serviceClusterIPRange := c.ServiceIPRange
	if serviceClusterIPRange.IP == nil {
		return LegacyRESTStorage{}, genericapiserver.APIGroupInfo{}, fmt.Errorf("service clusterIPRange is missing")
	}

	serviceStorageConfig, err := c.StorageFactory.NewConfig(api.Resource("services"))
	if err != nil {
		return LegacyRESTStorage{}, genericapiserver.APIGroupInfo{}, err
	}

	serviceClusterIPAllocator := ipallocator.NewAllocatorCIDRRange(&serviceClusterIPRange, func(max int, rangeSpec string) allocator.Interface {
		mem := allocator.NewAllocationMap(max, rangeSpec)
		// TODO etcdallocator package to return a storage interface via the storageFactory
		etcd := serviceallocator.NewEtcd(mem, "/ranges/serviceips", api.Resource("serviceipallocations"), serviceStorageConfig)
		serviceClusterIPRegistry = etcd
		return etcd
	})
	restStorage.ServiceClusterIPAllocator = serviceClusterIPRegistry

	var serviceNodePortRegistry rangeallocation.RangeRegistry
	serviceNodePortAllocator := portallocator.NewPortAllocatorCustom(c.ServiceNodePortRange, func(max int, rangeSpec string) allocator.Interface {
		mem := allocator.NewAllocationMap(max, rangeSpec)
		// TODO etcdallocator package to return a storage interface via the storageFactory
		etcd := serviceallocator.NewEtcd(mem, "/ranges/servicenodeports", api.Resource("servicenodeportallocations"), serviceStorageConfig)
		serviceNodePortRegistry = etcd
		return etcd
	})
	restStorage.ServiceNodePortAllocator = serviceNodePortRegistry

	controllerStorage := controllerstore.NewStorage(restOptionsGetter)

	serviceRest, serviceRestProxy := servicestore.NewREST(serviceRESTStorage, endpointsStorage, podStorage.Pod, serviceClusterIPAllocator, serviceNodePortAllocator, c.ProxyTransport)

	restStorageMap := map[string]rest.Storage{
		"pods":             podStorage.Pod,
		"pods/attach":      podStorage.Attach,
		"pods/status":      podStorage.Status,
		"pods/log":         podStorage.Log,
		"pods/exec":        podStorage.Exec,
		"pods/portforward": podStorage.PortForward,
		"pods/proxy":       podStorage.Proxy,
		"pods/binding":     podStorage.Binding,
		"bindings":         podStorage.Binding,

		"podTemplates": podTemplateStorage,

		"replicationControllers":        controllerStorage.Controller,
		"replicationControllers/status": controllerStorage.Status,

		"services":        serviceRest,
		"services/proxy":  serviceRestProxy,
		"services/status": serviceStatusStorage,

		"endpoints": endpointsStorage,

		"nodes":        nodeStorage.Node,
		"nodes/status": nodeStorage.Status,
		"nodes/proxy":  nodeStorage.Proxy,

		"events": eventStorage,

		"limitRanges":                   limitRangeStorage,
		"resourceQuotas":                resourceQuotaStorage,
		"resourceQuotas/status":         resourceQuotaStatusStorage,
		"namespaces":                    namespaceStorage,
		"namespaces/status":             namespaceStatusStorage,
		"namespaces/finalize":           namespaceFinalizeStorage,
		"secrets":                       secretStorage,
		"serviceAccounts":               serviceAccountStorage,
		"persistentVolumes":             persistentVolumeStorage,
		"persistentVolumes/status":      persistentVolumeStatusStorage,
		"persistentVolumeClaims":        persistentVolumeClaimStorage,
		"persistentVolumeClaims/status": persistentVolumeClaimStatusStorage,
		"configMaps":                    configMapStorage,

		"componentStatuses": componentstatus.NewStorage(componentStatusStorage{c.StorageFactory}.serversToValidate),
	}
	if legacyscheme.Scheme.IsVersionRegistered(schema.GroupVersion{Group: "autoscaling", Version: "v1"}) {
		restStorageMap["replicationControllers/scale"] = controllerStorage.Scale
	}
	if legacyscheme.Scheme.IsVersionRegistered(schema.GroupVersion{Group: "policy", Version: "v1beta1"}) {
		restStorageMap["pods/eviction"] = podStorage.Eviction
	}
	if serviceAccountStorage.Token != nil {
		restStorageMap["serviceaccounts/token"] = serviceAccountStorage.Token
	}
	apiGroupInfo.VersionedResourcesStorageMap["v1"] = restStorageMap

	return restStorage, apiGroupInfo, nil
}

func (p LegacyRESTStorageProvider) GroupName() string {
	return api.GroupName
}

type componentStatusStorage struct {
	storageFactory serverstorage.StorageFactory
}

func (s componentStatusStorage) serversToValidate() map[string]*componentstatus.Server {
	serversToValidate := map[string]*componentstatus.Server{
		"controller-manager": {Addr: "127.0.0.1", Port: ports.InsecureKubeControllerManagerPort, Path: "/healthz"},
		"scheduler":          {Addr: "127.0.0.1", Port: ports.InsecureSchedulerPort, Path: "/healthz"},
	}

	for ix, machine := range s.storageFactory.Backends() {
		etcdUrl, err := url.Parse(machine.Server)
		if err != nil {
			klog.Errorf("Failed to parse etcd url for validation: %v", err)
			continue
		}
		var port int
		var addr string
		if strings.Contains(etcdUrl.Host, ":") {
			var portString string
			addr, portString, err = net.SplitHostPort(etcdUrl.Host)
			if err != nil {
				klog.Errorf("Failed to split host/port: %s (%v)", etcdUrl.Host, err)
				continue
			}
			port, _ = strconv.Atoi(portString)
		} else {
			addr = etcdUrl.Host
			port = 2379
		}
		// TODO: etcd health checking should be abstracted in the storage tier
		serversToValidate[fmt.Sprintf("etcd-%d", ix)] = &componentstatus.Server{
			Addr:        addr,
			EnableHTTPS: etcdUrl.Scheme == "https",
			TLSConfig:   machine.TLSConfig,
			Port:        port,
			Path:        "/health",
			Validate:    etcdutil.EtcdHealthCheck,
		}
	}
	return serversToValidate
}
