/*
Copyright 2014 The Kubernetes Authors.

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
	"context"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/validation/field"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/storage/names"
	"github.com/aaron-prindle/krmapiserver/pkg/api/legacyscheme"
	resourcequotautil "github.com/aaron-prindle/krmapiserver/pkg/api/resourcequota"
	api "github.com/aaron-prindle/krmapiserver/pkg/apis/core"
	"github.com/aaron-prindle/krmapiserver/pkg/apis/core/validation"
)

// resourcequotaStrategy implements behavior for ResourceQuota objects
type resourcequotaStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating ResourceQuota
// objects via the REST API.
var Strategy = resourcequotaStrategy{legacyscheme.Scheme, names.SimpleNameGenerator}

// NamespaceScoped is true for resourcequotas.
func (resourcequotaStrategy) NamespaceScoped() bool {
	return true
}

// PrepareForCreate clears fields that are not allowed to be set by end users on creation.
func (resourcequotaStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	resourcequota := obj.(*api.ResourceQuota)
	resourcequota.Status = api.ResourceQuotaStatus{}
	resourcequotautil.DropDisabledFields(&resourcequota.Spec, nil)
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (resourcequotaStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newResourcequota := obj.(*api.ResourceQuota)
	oldResourcequota := old.(*api.ResourceQuota)
	newResourcequota.Status = oldResourcequota.Status
	resourcequotautil.DropDisabledFields(&newResourcequota.Spec, &oldResourcequota.Spec)
}

// Validate validates a new resourcequota.
func (resourcequotaStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	resourcequota := obj.(*api.ResourceQuota)
	return validation.ValidateResourceQuota(resourcequota)
}

// Canonicalize normalizes the object after validation.
func (resourcequotaStrategy) Canonicalize(obj runtime.Object) {
}

// AllowCreateOnUpdate is false for resourcequotas.
func (resourcequotaStrategy) AllowCreateOnUpdate() bool {
	return false
}

// ValidateUpdate is the default update validation for an end user.
func (resourcequotaStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	errorList := validation.ValidateResourceQuota(obj.(*api.ResourceQuota))
	return append(errorList, validation.ValidateResourceQuotaUpdate(obj.(*api.ResourceQuota), old.(*api.ResourceQuota))...)
}

func (resourcequotaStrategy) AllowUnconditionalUpdate() bool {
	return true
}

type resourcequotaStatusStrategy struct {
	resourcequotaStrategy
}

// StatusStrategy is the default logic invoked when updating object status.
var StatusStrategy = resourcequotaStatusStrategy{Strategy}

func (resourcequotaStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newResourcequota := obj.(*api.ResourceQuota)
	oldResourcequota := old.(*api.ResourceQuota)
	newResourcequota.Spec = oldResourcequota.Spec
}

func (resourcequotaStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateResourceQuotaStatusUpdate(obj.(*api.ResourceQuota), old.(*api.ResourceQuota))
}
