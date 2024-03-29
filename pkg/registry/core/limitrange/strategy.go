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

package limitrange

import (
	"context"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/uuid"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/validation/field"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/storage/names"
	"github.com/aaron-prindle/krmapiserver/pkg/api/legacyscheme"
	api "github.com/aaron-prindle/krmapiserver/pkg/apis/core"
	"github.com/aaron-prindle/krmapiserver/pkg/apis/core/validation"
)

type limitrangeStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating
// LimitRange objects via the REST API.
var Strategy = limitrangeStrategy{legacyscheme.Scheme, names.SimpleNameGenerator}

func (limitrangeStrategy) NamespaceScoped() bool {
	return true
}

func (limitrangeStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	limitRange := obj.(*api.LimitRange)
	if len(limitRange.Name) == 0 {
		limitRange.Name = string(uuid.NewUUID())
	}
}

func (limitrangeStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (limitrangeStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	limitRange := obj.(*api.LimitRange)
	return validation.ValidateLimitRange(limitRange)
}

// Canonicalize normalizes the object after validation.
func (limitrangeStrategy) Canonicalize(obj runtime.Object) {
}

func (limitrangeStrategy) AllowCreateOnUpdate() bool {
	return true
}

func (limitrangeStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	limitRange := obj.(*api.LimitRange)
	return validation.ValidateLimitRange(limitRange)
}

func (limitrangeStrategy) AllowUnconditionalUpdate() bool {
	return true
}

func (limitrangeStrategy) Export(context.Context, runtime.Object, bool) error {
	// Copied from OpenShift exporter
	// TODO: this needs to be fixed
	//  limitrange.Strategy.PrepareForCreate(ctx, obj)
	return nil
}
