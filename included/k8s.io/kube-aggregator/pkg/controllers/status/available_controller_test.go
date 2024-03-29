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

package apiserver

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/aaron-prindle/krmapiserver/included/github.com/davecgh/go-spew/spew"

	v1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/core/v1"
	metav1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/apis/meta/v1"
	v1listers "github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/listers/core/v1"
	clienttesting "github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/testing"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/tools/cache"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/kube-aggregator/pkg/apis/apiregistration"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/kube-aggregator/pkg/client/clientset_generated/internalclientset/fake"
	listers "github.com/aaron-prindle/krmapiserver/included/k8s.io/kube-aggregator/pkg/client/listers/apiregistration/internalversion"
)

const (
	testServicePort     = 1234
	testServicePortName = "testPort"
)

func newEndpoints(namespace, name string) *v1.Endpoints {
	return &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
	}
}

func newEndpointsWithAddress(namespace, name string, port int32, portName string) *v1.Endpoints {
	return &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP: "val",
					},
				},
				Ports: []v1.EndpointPort{
					{
						Name: portName,
						Port: port,
					},
				},
			},
		},
	}
}

func newService(namespace, name string, port int32, portName string) *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeClusterIP,
			Ports: []v1.ServicePort{
				{Port: port, Name: portName},
			},
		},
	}
}

func newLocalAPIService(name string) *apiregistration.APIService {
	return &apiregistration.APIService{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
}

func newRemoteAPIService(name string) *apiregistration.APIService {
	return &apiregistration.APIService{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: apiregistration.APIServiceSpec{
			Service: &apiregistration.ServiceReference{
				Namespace: "foo",
				Name:      "bar",
				Port:      testServicePort,
			},
		},
	}
}

func TestSync(t *testing.T) {
	tests := []struct {
		name string

		apiServiceName     string
		apiServices        []*apiregistration.APIService
		services           []*v1.Service
		endpoints          []*v1.Endpoints
		forceDiscoveryFail bool

		expectedAvailability apiregistration.APIServiceCondition
	}{
		{
			name:           "local",
			apiServiceName: "local.group",
			apiServices:    []*apiregistration.APIService{newLocalAPIService("local.group")},
			expectedAvailability: apiregistration.APIServiceCondition{
				Type:    apiregistration.Available,
				Status:  apiregistration.ConditionTrue,
				Reason:  "Local",
				Message: "Local APIServices are always available",
			},
		},
		{
			name:           "no service",
			apiServiceName: "remote.group",
			apiServices:    []*apiregistration.APIService{newRemoteAPIService("remote.group")},
			services:       []*v1.Service{newService("foo", "not-bar", testServicePort, testServicePortName)},
			expectedAvailability: apiregistration.APIServiceCondition{
				Type:    apiregistration.Available,
				Status:  apiregistration.ConditionFalse,
				Reason:  "ServiceNotFound",
				Message: `service/bar in "foo" is not present`,
			},
		},
		{
			name:           "service on bad port",
			apiServiceName: "remote.group",
			apiServices:    []*apiregistration.APIService{newRemoteAPIService("remote.group")},
			services: []*v1.Service{{
				ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"},
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeClusterIP,
					Ports: []v1.ServicePort{
						{Port: 6443},
					},
				},
			}},
			endpoints: []*v1.Endpoints{newEndpointsWithAddress("foo", "bar", testServicePort, testServicePortName)},
			expectedAvailability: apiregistration.APIServiceCondition{
				Type:    apiregistration.Available,
				Status:  apiregistration.ConditionFalse,
				Reason:  "ServicePortError",
				Message: fmt.Sprintf(`service/bar in "foo" is not listening on port %d`, testServicePort),
			},
		},
		{
			name:           "no endpoints",
			apiServiceName: "remote.group",
			apiServices:    []*apiregistration.APIService{newRemoteAPIService("remote.group")},
			services:       []*v1.Service{newService("foo", "bar", testServicePort, testServicePortName)},
			expectedAvailability: apiregistration.APIServiceCondition{
				Type:    apiregistration.Available,
				Status:  apiregistration.ConditionFalse,
				Reason:  "EndpointsNotFound",
				Message: `cannot find endpoints for service/bar in "foo"`,
			},
		},
		{
			name:           "missing endpoints",
			apiServiceName: "remote.group",
			apiServices:    []*apiregistration.APIService{newRemoteAPIService("remote.group")},
			services:       []*v1.Service{newService("foo", "bar", testServicePort, testServicePortName)},
			endpoints:      []*v1.Endpoints{newEndpoints("foo", "bar")},
			expectedAvailability: apiregistration.APIServiceCondition{
				Type:    apiregistration.Available,
				Status:  apiregistration.ConditionFalse,
				Reason:  "MissingEndpoints",
				Message: `endpoints for service/bar in "foo" have no addresses with port name "testPort"`,
			},
		},
		{
			name:           "wrong endpoint port name",
			apiServiceName: "remote.group",
			apiServices:    []*apiregistration.APIService{newRemoteAPIService("remote.group")},
			services:       []*v1.Service{newService("foo", "bar", testServicePort, testServicePortName)},
			endpoints:      []*v1.Endpoints{newEndpointsWithAddress("foo", "bar", testServicePort, "wrongName")},
			expectedAvailability: apiregistration.APIServiceCondition{
				Type:    apiregistration.Available,
				Status:  apiregistration.ConditionFalse,
				Reason:  "MissingEndpoints",
				Message: fmt.Sprintf(`endpoints for service/bar in "foo" have no addresses with port name "%s"`, testServicePortName),
			},
		},
		{
			name:           "remote",
			apiServiceName: "remote.group",
			apiServices:    []*apiregistration.APIService{newRemoteAPIService("remote.group")},
			services:       []*v1.Service{newService("foo", "bar", testServicePort, testServicePortName)},
			endpoints:      []*v1.Endpoints{newEndpointsWithAddress("foo", "bar", testServicePort, testServicePortName)},
			expectedAvailability: apiregistration.APIServiceCondition{
				Type:    apiregistration.Available,
				Status:  apiregistration.ConditionTrue,
				Reason:  "Passed",
				Message: `all checks passed`,
			},
		},
		{
			name:               "remote-bad-return",
			apiServiceName:     "remote.group",
			apiServices:        []*apiregistration.APIService{newRemoteAPIService("remote.group")},
			services:           []*v1.Service{newService("foo", "bar", testServicePort, testServicePortName)},
			endpoints:          []*v1.Endpoints{newEndpointsWithAddress("foo", "bar", testServicePort, testServicePortName)},
			forceDiscoveryFail: true,
			expectedAvailability: apiregistration.APIServiceCondition{
				Type:    apiregistration.Available,
				Status:  apiregistration.ConditionFalse,
				Reason:  "FailedDiscoveryCheck",
				Message: `failing or missing response from`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fakeClient := fake.NewSimpleClientset()
			apiServiceIndexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
			serviceIndexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
			endpointsIndexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
			for _, obj := range tc.apiServices {
				apiServiceIndexer.Add(obj)
			}
			for _, obj := range tc.services {
				serviceIndexer.Add(obj)
			}
			for _, obj := range tc.endpoints {
				endpointsIndexer.Add(obj)
			}

			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !tc.forceDiscoveryFail {
					w.WriteHeader(http.StatusOK)
				}
				w.WriteHeader(http.StatusForbidden)
			}))
			defer testServer.Close()

			c := AvailableConditionController{
				apiServiceClient: fakeClient.Apiregistration(),
				apiServiceLister: listers.NewAPIServiceLister(apiServiceIndexer),
				serviceLister:    v1listers.NewServiceLister(serviceIndexer),
				endpointsLister:  v1listers.NewEndpointsLister(endpointsIndexer),
				discoveryClient:  testServer.Client(),
				serviceResolver:  &fakeServiceResolver{url: testServer.URL},
			}
			c.sync(tc.apiServiceName)

			// ought to have one action writing status
			if e, a := 1, len(fakeClient.Actions()); e != a {
				t.Fatalf("%v expected %v, got %v", tc.name, e, fakeClient.Actions())
			}

			action, ok := fakeClient.Actions()[0].(clienttesting.UpdateAction)
			if !ok {
				t.Fatalf("%v got %v", tc.name, ok)
			}

			if e, a := 1, len(action.GetObject().(*apiregistration.APIService).Status.Conditions); e != a {
				t.Fatalf("%v expected %v, got %v", tc.name, e, action.GetObject())
			}
			condition := action.GetObject().(*apiregistration.APIService).Status.Conditions[0]
			if e, a := tc.expectedAvailability.Type, condition.Type; e != a {
				t.Errorf("%v expected %v, got %#v", tc.name, e, condition)
			}
			if e, a := tc.expectedAvailability.Status, condition.Status; e != a {
				t.Errorf("%v expected %v, got %#v", tc.name, e, condition)
			}
			if e, a := tc.expectedAvailability.Reason, condition.Reason; e != a {
				t.Errorf("%v expected %v, got %#v", tc.name, e, condition)
			}
			if e, a := tc.expectedAvailability.Message, condition.Message; !strings.HasPrefix(a, e) {
				t.Errorf("%v expected %v, got %#v", tc.name, e, condition)
			}
			if condition.LastTransitionTime.IsZero() {
				t.Error("expected lastTransitionTime to be non-zero")
			}
		})
	}
}

type fakeServiceResolver struct {
	url string
}

func (f *fakeServiceResolver) ResolveEndpoint(namespace, name string, port int32) (*url.URL, error) {
	return url.Parse(f.url)
}

func TestUpdateAPIServiceStatus(t *testing.T) {
	foo := &apiregistration.APIService{Status: apiregistration.APIServiceStatus{Conditions: []apiregistration.APIServiceCondition{{Type: "foo"}}}}
	bar := &apiregistration.APIService{Status: apiregistration.APIServiceStatus{Conditions: []apiregistration.APIServiceCondition{{Type: "bar"}}}}

	fakeClient := fake.NewSimpleClientset()
	updateAPIServiceStatus(fakeClient.Apiregistration(), foo, foo)
	if e, a := 0, len(fakeClient.Actions()); e != a {
		t.Error(spew.Sdump(fakeClient.Actions()))
	}

	fakeClient.ClearActions()
	updateAPIServiceStatus(fakeClient.Apiregistration(), foo, bar)
	if e, a := 1, len(fakeClient.Actions()); e != a {
		t.Error(spew.Sdump(fakeClient.Actions()))
	}

}
