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

package capabilities

import (
	"reflect"
	"testing"

	corev1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/core/v1"
	policy "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/policy/v1beta1"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/validation/field"
	api "github.com/aaron-prindle/krmapiserver/pkg/apis/core"
)

func TestGenerateAdds(t *testing.T) {
	tests := map[string]struct {
		defaultAddCaps []corev1.Capability
		containerCaps  *api.Capabilities
		expectedCaps   *api.Capabilities
	}{
		"no required, no container requests": {},
		"no required, no container requests, non-nil": {
			containerCaps: &api.Capabilities{},
			expectedCaps:  &api.Capabilities{},
		},
		"required, no container requests": {
			defaultAddCaps: []corev1.Capability{"foo"},
			expectedCaps: &api.Capabilities{
				Add: []api.Capability{"foo"},
			},
		},
		"required, container requests add required": {
			defaultAddCaps: []corev1.Capability{"foo"},
			containerCaps: &api.Capabilities{
				Add: []api.Capability{"foo"},
			},
			expectedCaps: &api.Capabilities{
				Add: []api.Capability{"foo"},
			},
		},
		"multiple required, container requests add required": {
			defaultAddCaps: []corev1.Capability{"foo", "bar", "baz"},
			containerCaps: &api.Capabilities{
				Add: []api.Capability{"foo"},
			},
			expectedCaps: &api.Capabilities{
				Add: []api.Capability{"bar", "baz", "foo"},
			},
		},
		"required, container requests add non-required": {
			defaultAddCaps: []corev1.Capability{"foo"},
			containerCaps: &api.Capabilities{
				Add: []api.Capability{"bar"},
			},
			expectedCaps: &api.Capabilities{
				Add: []api.Capability{"bar", "foo"},
			},
		},
		"generation does not mutate unnecessarily": {
			defaultAddCaps: []corev1.Capability{"foo", "bar"},
			containerCaps: &api.Capabilities{
				Add: []api.Capability{"foo", "foo", "bar", "baz"},
			},
			expectedCaps: &api.Capabilities{
				Add: []api.Capability{"foo", "foo", "bar", "baz"},
			},
		},
		"generation dedupes": {
			defaultAddCaps: []corev1.Capability{"foo", "bar"},
			containerCaps: &api.Capabilities{
				Add: []api.Capability{"foo", "baz"},
			},
			expectedCaps: &api.Capabilities{
				Add: []api.Capability{"bar", "baz", "foo"},
			},
		},
		"generation is case sensitive - will not dedupe": {
			defaultAddCaps: []corev1.Capability{"foo"},
			containerCaps: &api.Capabilities{
				Add: []api.Capability{"FOO"},
			},
			expectedCaps: &api.Capabilities{
				Add: []api.Capability{"FOO", "foo"},
			},
		},
	}

	for k, v := range tests {
		container := &api.Container{
			SecurityContext: &api.SecurityContext{
				Capabilities: v.containerCaps,
			},
		}

		strategy, err := NewDefaultCapabilities(v.defaultAddCaps, nil, nil)
		if err != nil {
			t.Errorf("%s failed: %v", k, err)
			continue
		}
		generatedCaps, err := strategy.Generate(nil, container)
		if err != nil {
			t.Errorf("%s failed generating: %v", k, err)
			continue
		}
		if v.expectedCaps == nil && generatedCaps != nil {
			t.Errorf("%s expected nil caps to be generated but got %v", k, generatedCaps)
			continue
		}
		if !reflect.DeepEqual(v.expectedCaps, generatedCaps) {
			t.Errorf("%s did not generate correctly.  Expected: %#v, Actual: %#v", k, v.expectedCaps, generatedCaps)
		}
	}
}

func TestGenerateDrops(t *testing.T) {
	tests := map[string]struct {
		defaultAddCaps   []corev1.Capability
		requiredDropCaps []corev1.Capability
		containerCaps    *api.Capabilities
		expectedCaps     *api.Capabilities
	}{
		"no required, no container requests": {
			expectedCaps: nil,
		},
		"no required, no container requests, non-nil": {
			containerCaps: &api.Capabilities{},
			expectedCaps:  &api.Capabilities{},
		},
		"required drops are defaulted": {
			requiredDropCaps: []corev1.Capability{"foo"},
			expectedCaps: &api.Capabilities{
				Drop: []api.Capability{"foo"},
			},
		},
		"required drops are defaulted when making container requests": {
			requiredDropCaps: []corev1.Capability{"baz"},
			containerCaps: &api.Capabilities{
				Drop: []api.Capability{"foo", "bar"},
			},
			expectedCaps: &api.Capabilities{
				Drop: []api.Capability{"bar", "baz", "foo"},
			},
		},
		"required drops do not mutate unnecessarily": {
			requiredDropCaps: []corev1.Capability{"baz"},
			containerCaps: &api.Capabilities{
				Drop: []api.Capability{"foo", "bar", "baz"},
			},
			expectedCaps: &api.Capabilities{
				Drop: []api.Capability{"foo", "bar", "baz"},
			},
		},
		"can drop a required add": {
			defaultAddCaps: []corev1.Capability{"foo"},
			containerCaps: &api.Capabilities{
				Drop: []api.Capability{"foo"},
			},
			expectedCaps: &api.Capabilities{
				Drop: []api.Capability{"foo"},
			},
		},
		"can drop non-required add": {
			defaultAddCaps: []corev1.Capability{"foo"},
			containerCaps: &api.Capabilities{
				Drop: []api.Capability{"bar"},
			},
			expectedCaps: &api.Capabilities{
				Add:  []api.Capability{"foo"},
				Drop: []api.Capability{"bar"},
			},
		},
		"defaulting adds and drops, dropping a required add": {
			defaultAddCaps:   []corev1.Capability{"foo", "bar", "baz"},
			requiredDropCaps: []corev1.Capability{"abc"},
			containerCaps: &api.Capabilities{
				Drop: []api.Capability{"foo"},
			},
			expectedCaps: &api.Capabilities{
				Add:  []api.Capability{"bar", "baz"},
				Drop: []api.Capability{"abc", "foo"},
			},
		},
		"generation dedupes": {
			requiredDropCaps: []corev1.Capability{"baz", "foo"},
			containerCaps: &api.Capabilities{
				Drop: []api.Capability{"bar", "foo"},
			},
			expectedCaps: &api.Capabilities{
				Drop: []api.Capability{"bar", "baz", "foo"},
			},
		},
		"generation is case sensitive - will not dedupe": {
			requiredDropCaps: []corev1.Capability{"bar"},
			containerCaps: &api.Capabilities{
				Drop: []api.Capability{"BAR"},
			},
			expectedCaps: &api.Capabilities{
				Drop: []api.Capability{"BAR", "bar"},
			},
		},
	}
	for k, v := range tests {
		container := &api.Container{
			SecurityContext: &api.SecurityContext{
				Capabilities: v.containerCaps,
			},
		}

		strategy, err := NewDefaultCapabilities(v.defaultAddCaps, v.requiredDropCaps, nil)
		if err != nil {
			t.Errorf("%s failed: %v", k, err)
			continue
		}
		generatedCaps, err := strategy.Generate(nil, container)
		if err != nil {
			t.Errorf("%s failed generating: %v", k, err)
			continue
		}
		if v.expectedCaps == nil && generatedCaps != nil {
			t.Errorf("%s expected nil caps to be generated but got %#v", k, generatedCaps)
			continue
		}
		if !reflect.DeepEqual(v.expectedCaps, generatedCaps) {
			t.Errorf("%s did not generate correctly.  Expected: %#v, Actual: %#v", k, v.expectedCaps, generatedCaps)
		}
	}
}

func TestValidateAdds(t *testing.T) {
	tests := map[string]struct {
		defaultAddCaps []corev1.Capability
		allowedCaps    []corev1.Capability
		containerCaps  *api.Capabilities
		expectedError  string
	}{
		// no container requests
		"no required, no allowed, no container requests": {},
		"no required, allowed, no container requests": {
			allowedCaps: []corev1.Capability{"foo"},
		},
		"required, no allowed, no container requests": {
			defaultAddCaps: []corev1.Capability{"foo"},
			expectedError:  `capabilities: Invalid value: "null": required capabilities are not set on the securityContext`,
		},

		// container requests match required
		"required, no allowed, container requests valid": {
			defaultAddCaps: []corev1.Capability{"foo"},
			containerCaps: &api.Capabilities{
				Add: []api.Capability{"foo"},
			},
		},
		"required, no allowed, container requests invalid": {
			defaultAddCaps: []corev1.Capability{"foo"},
			containerCaps: &api.Capabilities{
				Add: []api.Capability{"bar"},
			},
			expectedError: `capabilities.add: Invalid value: "bar": capability may not be added`,
		},

		// container requests match allowed
		"no required, allowed, container requests valid": {
			allowedCaps: []corev1.Capability{"foo"},
			containerCaps: &api.Capabilities{
				Add: []api.Capability{"foo"},
			},
		},
		"no required, all allowed, container requests valid": {
			allowedCaps: []corev1.Capability{policy.AllowAllCapabilities},
			containerCaps: &api.Capabilities{
				Add: []api.Capability{"foo"},
			},
		},
		"no required, allowed, container requests invalid": {
			allowedCaps: []corev1.Capability{"foo"},
			containerCaps: &api.Capabilities{
				Add: []api.Capability{"bar"},
			},
			expectedError: `capabilities.add: Invalid value: "bar": capability may not be added`,
		},

		// required and allowed
		"required, allowed, container requests valid required": {
			defaultAddCaps: []corev1.Capability{"foo"},
			allowedCaps:    []corev1.Capability{"bar"},
			containerCaps: &api.Capabilities{
				Add: []api.Capability{"foo"},
			},
		},
		"required, allowed, container requests valid allowed": {
			defaultAddCaps: []corev1.Capability{"foo"},
			allowedCaps:    []corev1.Capability{"bar"},
			containerCaps: &api.Capabilities{
				Add: []api.Capability{"bar"},
			},
		},
		"required, allowed, container requests invalid": {
			defaultAddCaps: []corev1.Capability{"foo"},
			allowedCaps:    []corev1.Capability{"bar"},
			containerCaps: &api.Capabilities{
				Add: []api.Capability{"baz"},
			},
			expectedError: `capabilities.add: Invalid value: "baz": capability may not be added`,
		},
		"validation is case sensitive": {
			defaultAddCaps: []corev1.Capability{"foo"},
			containerCaps: &api.Capabilities{
				Add: []api.Capability{"FOO"},
			},
			expectedError: `capabilities.add: Invalid value: "FOO": capability may not be added`,
		},
	}

	for k, v := range tests {
		strategy, err := NewDefaultCapabilities(v.defaultAddCaps, nil, v.allowedCaps)
		if err != nil {
			t.Errorf("%s failed: %v", k, err)
			continue
		}
		errs := strategy.Validate(field.NewPath("capabilities"), nil, nil, v.containerCaps)
		if v.expectedError == "" && len(errs) > 0 {
			t.Errorf("%s should have passed but had errors %v", k, errs)
			continue
		}
		if v.expectedError != "" && len(errs) == 0 {
			t.Errorf("%s should have failed but received no errors", k)
			continue
		}
		if len(errs) == 1 && errs[0].Error() != v.expectedError {
			t.Errorf("%s should have failed with %v but received %v", k, v.expectedError, errs[0])
			continue
		}
		if len(errs) > 1 {
			t.Errorf("%s should have failed with at most one error, but received %v: %v", k, len(errs), errs)
		}
	}
}

func TestValidateDrops(t *testing.T) {
	tests := map[string]struct {
		requiredDropCaps []corev1.Capability
		containerCaps    *api.Capabilities
		expectedError    string
	}{
		// no container requests
		"no required, no container requests": {},
		"required, no container requests": {
			requiredDropCaps: []corev1.Capability{"foo"},
			expectedError:    `capabilities: Invalid value: "null": required capabilities are not set on the securityContext`,
		},

		// container requests match required
		"required, container requests valid": {
			requiredDropCaps: []corev1.Capability{"foo"},
			containerCaps: &api.Capabilities{
				Drop: []api.Capability{"foo"},
			},
		},
		"required, container requests invalid": {
			requiredDropCaps: []corev1.Capability{"foo"},
			containerCaps: &api.Capabilities{
				Drop: []api.Capability{"bar"},
			},
			expectedError: `capabilities.drop: Invalid value: []core.Capability{"bar"}: foo is required to be dropped but was not found`,
		},
		"validation is case sensitive": {
			requiredDropCaps: []corev1.Capability{"foo"},
			containerCaps: &api.Capabilities{
				Drop: []api.Capability{"FOO"},
			},
			expectedError: `capabilities.drop: Invalid value: []core.Capability{"FOO"}: foo is required to be dropped but was not found`,
		},
	}

	for k, v := range tests {
		strategy, err := NewDefaultCapabilities(nil, v.requiredDropCaps, nil)
		if err != nil {
			t.Errorf("%s failed: %v", k, err)
			continue
		}
		errs := strategy.Validate(field.NewPath("capabilities"), nil, nil, v.containerCaps)
		if v.expectedError == "" && len(errs) > 0 {
			t.Errorf("%s should have passed but had errors %v", k, errs)
			continue
		}
		if v.expectedError != "" && len(errs) == 0 {
			t.Errorf("%s should have failed but received no errors", k)
			continue
		}
		if len(errs) == 1 && errs[0].Error() != v.expectedError {
			t.Errorf("%s should have failed with %v but received %v", k, v.expectedError, errs[0])
			continue
		}
		if len(errs) > 1 {
			t.Errorf("%s should have failed with at most one error, but received %v: %v", k, len(errs), errs)
		}
	}
}
