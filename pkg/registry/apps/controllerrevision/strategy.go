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

package controllerrevision

import (
	"context"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/validation/field"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/rest"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/storage/names"
	"github.com/aaron-prindle/krmapiserver/pkg/api/legacyscheme"
	"github.com/aaron-prindle/krmapiserver/pkg/apis/apps"
	"github.com/aaron-prindle/krmapiserver/pkg/apis/apps/validation"
)

// strategy implements behavior for ConfigMap objects
type strategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating ControllerRevision
// objects via the REST API.
var Strategy = strategy{legacyscheme.Scheme, names.SimpleNameGenerator}

// Strategy should implement rest.RESTCreateStrategy
var _ rest.RESTCreateStrategy = Strategy

// Strategy should implement rest.RESTUpdateStrategy
var _ rest.RESTUpdateStrategy = Strategy

func (strategy) NamespaceScoped() bool {
	return true
}

func (strategy) Canonicalize(obj runtime.Object) {
}

func (strategy) AllowCreateOnUpdate() bool {
	return false
}

func (strategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	_ = obj.(*apps.ControllerRevision)
}

func (strategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	revision := obj.(*apps.ControllerRevision)

	return validation.ValidateControllerRevision(revision)
}

func (strategy) PrepareForUpdate(ctx context.Context, newObj, oldObj runtime.Object) {
	_ = oldObj.(*apps.ControllerRevision)
	_ = newObj.(*apps.ControllerRevision)
}

func (strategy) AllowUnconditionalUpdate() bool {
	return true
}

func (strategy) ValidateUpdate(ctx context.Context, newObj, oldObj runtime.Object) field.ErrorList {
	oldRevision, newRevision := oldObj.(*apps.ControllerRevision), newObj.(*apps.ControllerRevision)
	return validation.ValidateControllerRevisionUpdate(newRevision, oldRevision)
}
