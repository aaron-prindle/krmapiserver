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

package lease

import (
	"context"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/validation/field"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/storage/names"
	"github.com/aaron-prindle/krmapiserver/pkg/api/legacyscheme"
	"github.com/aaron-prindle/krmapiserver/pkg/apis/coordination"
	"github.com/aaron-prindle/krmapiserver/pkg/apis/coordination/validation"
)

// leaseStrategy implements verification logic for Leases.
type leaseStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating Lease objects.
var Strategy = leaseStrategy{legacyscheme.Scheme, names.SimpleNameGenerator}

// NamespaceScoped returns true because all Lease' need to be within a namespace.
func (leaseStrategy) NamespaceScoped() bool {
	return true
}

// PrepareForCreate prepares Lease for creation.
func (leaseStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (leaseStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

// Validate validates a new Lease.
func (leaseStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	lease := obj.(*coordination.Lease)
	return validation.ValidateLease(lease)
}

// Canonicalize normalizes the object after validation.
func (leaseStrategy) Canonicalize(obj runtime.Object) {
}

// AllowCreateOnUpdate is true for Lease; this means you may create one with a PUT request.
func (leaseStrategy) AllowCreateOnUpdate() bool {
	return true
}

// ValidateUpdate is the default update validation for an end user.
func (leaseStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateLeaseUpdate(obj.(*coordination.Lease), old.(*coordination.Lease))
}

// AllowUnconditionalUpdate is the default update policy for Lease objects.
func (leaseStrategy) AllowUnconditionalUpdate() bool {
	return false
}
