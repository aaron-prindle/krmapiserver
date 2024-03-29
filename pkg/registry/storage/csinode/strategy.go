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

package csinode

import (
	"context"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/validation/field"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/storage/names"
	"github.com/aaron-prindle/krmapiserver/pkg/api/legacyscheme"
	"github.com/aaron-prindle/krmapiserver/pkg/apis/storage"
	"github.com/aaron-prindle/krmapiserver/pkg/apis/storage/validation"
)

// csiNodeStrategy implements behavior for CSINode objects
type csiNodeStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating
// CSINode objects via the REST API.
var Strategy = csiNodeStrategy{legacyscheme.Scheme, names.SimpleNameGenerator}

func (csiNodeStrategy) NamespaceScoped() bool {
	return false
}

// ResetBeforeCreate clears the Status field which is not allowed to be set by end users on creation.
func (csiNodeStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (csiNodeStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	csiNode := obj.(*storage.CSINode)

	errs := validation.ValidateCSINode(csiNode)
	errs = append(errs, validation.ValidateCSINode(csiNode)...)

	return errs
}

// Canonicalize normalizes the object after validation.
func (csiNodeStrategy) Canonicalize(obj runtime.Object) {
}

func (csiNodeStrategy) AllowCreateOnUpdate() bool {
	return false
}

// PrepareForUpdate sets the Status fields which is not allowed to be set by an end user updating a CSINode
func (csiNodeStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (csiNodeStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newCSINodeObj := obj.(*storage.CSINode)
	oldCSINodeObj := old.(*storage.CSINode)
	errorList := validation.ValidateCSINode(newCSINodeObj)
	return append(errorList, validation.ValidateCSINodeUpdate(newCSINodeObj, oldCSINodeObj)...)
}

func (csiNodeStrategy) AllowUnconditionalUpdate() bool {
	return false
}
