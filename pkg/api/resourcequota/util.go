/*
Copyright 2019 The Kubernetes Authors.

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

package resourcequota

import (
	utilfeature "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/util/feature"
	api "github.com/aaron-prindle/krmapiserver/pkg/apis/core"
	"github.com/aaron-prindle/krmapiserver/pkg/features"
)

// DropDisabledFields removes disabled fields from the ResourceQuota spec.
// This should be called from PrepareForCreate/PrepareForUpdate for all resources containing a ResourceQuota spec.
func DropDisabledFields(resSpec *api.ResourceQuotaSpec, oldResSpec *api.ResourceQuotaSpec) {
	if !utilfeature.DefaultFeatureGate.Enabled(features.ResourceQuotaScopeSelectors) && !resourceQuotaScopeSelectorInUse(oldResSpec) {
		resSpec.ScopeSelector = nil
	}
}

func resourceQuotaScopeSelectorInUse(oldResSpec *api.ResourceQuotaSpec) bool {
	if oldResSpec == nil {
		return false
	}
	if oldResSpec.ScopeSelector != nil {
		return true
	}
	return false
}
