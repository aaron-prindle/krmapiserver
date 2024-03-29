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

package podsecuritypolicy

import (
	policy "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/policy/v1beta1"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/validation/field"
	api "github.com/aaron-prindle/krmapiserver/pkg/apis/core"
	"github.com/aaron-prindle/krmapiserver/pkg/security/podsecuritypolicy/apparmor"
	"github.com/aaron-prindle/krmapiserver/pkg/security/podsecuritypolicy/capabilities"
	"github.com/aaron-prindle/krmapiserver/pkg/security/podsecuritypolicy/group"
	"github.com/aaron-prindle/krmapiserver/pkg/security/podsecuritypolicy/seccomp"
	"github.com/aaron-prindle/krmapiserver/pkg/security/podsecuritypolicy/selinux"
	"github.com/aaron-prindle/krmapiserver/pkg/security/podsecuritypolicy/sysctl"
	"github.com/aaron-prindle/krmapiserver/pkg/security/podsecuritypolicy/user"
)

// Provider provides the implementation to generate a new security
// context based on constraints or validate an existing security context against constraints.
type Provider interface {
	// MutatePod sets the default values of the required but not filled fields of the pod and all
	// containers in the pod.
	MutatePod(pod *api.Pod) error
	// ValidatePod ensures a pod and all its containers are in compliance with the given constraints.
	// ValidatePod MUST NOT mutate the pod.
	ValidatePod(pod *api.Pod) field.ErrorList
	// Get the name of the PSP that this provider was initialized with.
	GetPSPName() string
}

// StrategyFactory abstracts how the strategies are created from the provider so that you may
// implement your own custom strategies that may pull information from other resources as necessary.
// For example, if you would like to populate the strategies with values from namespace annotations
// you may create a factory with a client that can pull the namespace and populate the appropriate
// values.
type StrategyFactory interface {
	// CreateStrategies creates the strategies that a provider will use.  The namespace argument
	// should be the namespace of the object being checked (the pod's namespace).
	CreateStrategies(psp *policy.PodSecurityPolicy, namespace string) (*ProviderStrategies, error)
}

// ProviderStrategies is a holder for all strategies that the provider requires to be populated.
type ProviderStrategies struct {
	RunAsUserStrategy         user.RunAsUserStrategy
	RunAsGroupStrategy        group.GroupStrategy
	SELinuxStrategy           selinux.SELinuxStrategy
	AppArmorStrategy          apparmor.Strategy
	FSGroupStrategy           group.GroupStrategy
	SupplementalGroupStrategy group.GroupStrategy
	CapabilitiesStrategy      capabilities.Strategy
	SysctlsStrategy           sysctl.SysctlsStrategy
	SeccompStrategy           seccomp.Strategy
}
