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

package service

import (
	"strings"
	"testing"

	api "github.com/aaron-prindle/krmapiserver/pkg/apis/core"
	utilnet "github.com/aaron-prindle/krmapiserver/included/k8s.io/utils/net"
)

func TestGetLoadBalancerSourceRanges(t *testing.T) {
	checkError := func(v string) {
		annotations := make(map[string]string)
		annotations[api.AnnotationLoadBalancerSourceRangesKey] = v
		svc := api.Service{}
		svc.Annotations = annotations
		_, err := GetLoadBalancerSourceRanges(&svc)
		if err == nil {
			t.Errorf("Expected error parsing: %q", v)
		}
		svc = api.Service{}
		svc.Spec.LoadBalancerSourceRanges = strings.Split(v, ",")
		_, err = GetLoadBalancerSourceRanges(&svc)
		if err == nil {
			t.Errorf("Expected error parsing: %q", v)
		}
	}
	checkError("10.0.0.1/33")
	checkError("foo.bar")
	checkError("10.0.0.1/32,*")
	checkError("10.0.0.1/32,")
	checkError("10.0.0.1/32, ")
	checkError("10.0.0.1")

	checkOK := func(v string) utilnet.IPNetSet {
		annotations := make(map[string]string)
		annotations[api.AnnotationLoadBalancerSourceRangesKey] = v
		svc := api.Service{}
		svc.Annotations = annotations
		_, err := GetLoadBalancerSourceRanges(&svc)
		if err != nil {
			t.Errorf("Unexpected error parsing: %q", v)
		}
		svc = api.Service{}
		svc.Spec.LoadBalancerSourceRanges = strings.Split(v, ",")
		cidrs, err := GetLoadBalancerSourceRanges(&svc)
		if err != nil {
			t.Errorf("Unexpected error parsing: %q", v)
		}
		return cidrs
	}
	cidrs := checkOK("192.168.0.1/32")
	if len(cidrs) != 1 {
		t.Errorf("Expected exactly one CIDR: %v", cidrs.StringSlice())
	}
	cidrs = checkOK("192.168.0.1/32,192.168.0.1/32")
	if len(cidrs) != 1 {
		t.Errorf("Expected exactly one CIDR (after de-dup): %v", cidrs.StringSlice())
	}
	cidrs = checkOK("192.168.0.1/32,192.168.0.2/32")
	if len(cidrs) != 2 {
		t.Errorf("Expected two CIDRs: %v", cidrs.StringSlice())
	}
	cidrs = checkOK("  192.168.0.1/32 , 192.168.0.2/32   ")
	if len(cidrs) != 2 {
		t.Errorf("Expected two CIDRs: %v", cidrs.StringSlice())
	}
	// check LoadBalancerSourceRanges not specified
	svc := api.Service{}
	cidrs, err := GetLoadBalancerSourceRanges(&svc)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(cidrs) != 1 {
		t.Errorf("Expected exactly one CIDR: %v", cidrs.StringSlice())
	}
	if !IsAllowAll(cidrs) {
		t.Errorf("Expected default to be allow-all: %v", cidrs.StringSlice())
	}
	// check SourceRanges annotation is empty
	annotations := make(map[string]string)
	annotations[api.AnnotationLoadBalancerSourceRangesKey] = ""
	svc = api.Service{}
	svc.Annotations = annotations
	cidrs, err = GetLoadBalancerSourceRanges(&svc)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(cidrs) != 1 {
		t.Errorf("Expected exactly one CIDR: %v", cidrs.StringSlice())
	}
	if !IsAllowAll(cidrs) {
		t.Errorf("Expected default to be allow-all: %v", cidrs.StringSlice())
	}
}

func TestAllowAll(t *testing.T) {
	checkAllowAll := func(allowAll bool, cidrs ...string) {
		ipnets, err := utilnet.ParseIPNets(cidrs...)
		if err != nil {
			t.Errorf("Unexpected error parsing cidrs: %v", cidrs)
		}
		if allowAll != IsAllowAll(ipnets) {
			t.Errorf("IsAllowAll did not return expected value for %v", cidrs)
		}
	}
	checkAllowAll(false, "10.0.0.1/32")
	checkAllowAll(false, "10.0.0.1/32", "10.0.0.2/32")
	checkAllowAll(false, "10.0.0.1/32", "10.0.0.1/32")

	checkAllowAll(true, "0.0.0.0/0")
	checkAllowAll(true, "192.168.0.0/0")
	checkAllowAll(true, "192.168.0.1/32", "0.0.0.0/0")
}

func TestRequestsOnlyLocalTraffic(t *testing.T) {
	checkRequestsOnlyLocalTraffic := func(requestsOnlyLocalTraffic bool, service *api.Service) {
		res := RequestsOnlyLocalTraffic(service)
		if res != requestsOnlyLocalTraffic {
			t.Errorf("Expected requests OnlyLocal traffic = %v, got %v",
				requestsOnlyLocalTraffic, res)
		}
	}

	checkRequestsOnlyLocalTraffic(false, &api.Service{})
	checkRequestsOnlyLocalTraffic(false, &api.Service{
		Spec: api.ServiceSpec{
			Type: api.ServiceTypeClusterIP,
		},
	})
	checkRequestsOnlyLocalTraffic(false, &api.Service{
		Spec: api.ServiceSpec{
			Type: api.ServiceTypeNodePort,
		},
	})
	checkRequestsOnlyLocalTraffic(false, &api.Service{
		Spec: api.ServiceSpec{
			Type:                  api.ServiceTypeNodePort,
			ExternalTrafficPolicy: api.ServiceExternalTrafficPolicyTypeCluster,
		},
	})
	checkRequestsOnlyLocalTraffic(true, &api.Service{
		Spec: api.ServiceSpec{
			Type:                  api.ServiceTypeNodePort,
			ExternalTrafficPolicy: api.ServiceExternalTrafficPolicyTypeLocal,
		},
	})
	checkRequestsOnlyLocalTraffic(false, &api.Service{
		Spec: api.ServiceSpec{
			Type:                  api.ServiceTypeLoadBalancer,
			ExternalTrafficPolicy: api.ServiceExternalTrafficPolicyTypeCluster,
		},
	})
	checkRequestsOnlyLocalTraffic(true, &api.Service{
		Spec: api.ServiceSpec{
			Type:                  api.ServiceTypeLoadBalancer,
			ExternalTrafficPolicy: api.ServiceExternalTrafficPolicyTypeLocal,
		},
	})
}

func TestNeedsHealthCheck(t *testing.T) {
	checkNeedsHealthCheck := func(needsHealthCheck bool, service *api.Service) {
		res := NeedsHealthCheck(service)
		if res != needsHealthCheck {
			t.Errorf("Expected needs health check = %v, got %v",
				needsHealthCheck, res)
		}
	}

	checkNeedsHealthCheck(false, &api.Service{
		Spec: api.ServiceSpec{
			Type: api.ServiceTypeClusterIP,
		},
	})
	checkNeedsHealthCheck(false, &api.Service{
		Spec: api.ServiceSpec{
			Type:                  api.ServiceTypeNodePort,
			ExternalTrafficPolicy: api.ServiceExternalTrafficPolicyTypeCluster,
		},
	})
	checkNeedsHealthCheck(false, &api.Service{
		Spec: api.ServiceSpec{
			Type:                  api.ServiceTypeNodePort,
			ExternalTrafficPolicy: api.ServiceExternalTrafficPolicyTypeLocal,
		},
	})
	checkNeedsHealthCheck(false, &api.Service{
		Spec: api.ServiceSpec{
			Type:                  api.ServiceTypeLoadBalancer,
			ExternalTrafficPolicy: api.ServiceExternalTrafficPolicyTypeCluster,
		},
	})
	checkNeedsHealthCheck(true, &api.Service{
		Spec: api.ServiceSpec{
			Type:                  api.ServiceTypeLoadBalancer,
			ExternalTrafficPolicy: api.ServiceExternalTrafficPolicyTypeLocal,
		},
	})
}
