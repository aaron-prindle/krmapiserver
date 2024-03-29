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

package endpoint

import (
	"context"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/validation/field"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/storage/names"
	endptspkg "github.com/aaron-prindle/krmapiserver/pkg/api/endpoints"
	"github.com/aaron-prindle/krmapiserver/pkg/api/legacyscheme"
	api "github.com/aaron-prindle/krmapiserver/pkg/apis/core"
	"github.com/aaron-prindle/krmapiserver/pkg/apis/core/validation"
)

// endpointsStrategy implements behavior for Endpoints
type endpointsStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating Endpoint
// objects via the REST API.
var Strategy = endpointsStrategy{legacyscheme.Scheme, names.SimpleNameGenerator}

// NamespaceScoped is true for endpoints.
func (endpointsStrategy) NamespaceScoped() bool {
	return true
}

// PrepareForCreate clears fields that are not allowed to be set by end users on creation.
func (endpointsStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (endpointsStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

// Validate validates a new endpoints.
func (endpointsStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	allErrs := validation.ValidateEndpoints(obj.(*api.Endpoints))
	allErrs = append(allErrs, validation.ValidateConditionalEndpoints(obj.(*api.Endpoints), nil)...)
	return allErrs
}

// Canonicalize normalizes the object after validation.
func (endpointsStrategy) Canonicalize(obj runtime.Object) {
	endpoints := obj.(*api.Endpoints)
	endpoints.Subsets = endptspkg.RepackSubsets(endpoints.Subsets)
}

// AllowCreateOnUpdate is true for endpoints.
func (endpointsStrategy) AllowCreateOnUpdate() bool {
	return true
}

// ValidateUpdate is the default update validation for an end user.
func (endpointsStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	errorList := validation.ValidateEndpoints(obj.(*api.Endpoints))
	errorList = append(errorList, validation.ValidateEndpointsUpdate(obj.(*api.Endpoints), old.(*api.Endpoints))...)
	errorList = append(errorList, validation.ValidateConditionalEndpoints(obj.(*api.Endpoints), old.(*api.Endpoints))...)
	return errorList
}

func (endpointsStrategy) AllowUnconditionalUpdate() bool {
	return true
}
