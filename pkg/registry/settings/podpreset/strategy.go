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

package podpreset

import (
	"context"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/validation/field"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/storage/names"
	"github.com/aaron-prindle/krmapiserver/pkg/api/legacyscheme"
	"github.com/aaron-prindle/krmapiserver/pkg/apis/settings"
	"github.com/aaron-prindle/krmapiserver/pkg/apis/settings/validation"
)

// podPresetStrategy implements verification logic for Pod Presets.
type podPresetStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating Pod Preset objects.
var Strategy = podPresetStrategy{legacyscheme.Scheme, names.SimpleNameGenerator}

// NamespaceScoped returns true because all Pod Presets need to be within a namespace.
func (podPresetStrategy) NamespaceScoped() bool {
	return true
}

// PrepareForCreate clears the status of a Pod Preset before creation.
func (podPresetStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	pip := obj.(*settings.PodPreset)
	pip.Generation = 1
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (podPresetStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newPodPreset := obj.(*settings.PodPreset)
	oldPodPreset := old.(*settings.PodPreset)

	// Update is not allowed
	newPodPreset.Spec = oldPodPreset.Spec
}

// Validate validates a new PodPreset.
func (podPresetStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	pip := obj.(*settings.PodPreset)
	return validation.ValidatePodPreset(pip)
}

// Canonicalize normalizes the object after validation.
func (podPresetStrategy) Canonicalize(obj runtime.Object) {}

// AllowCreateOnUpdate is false for PodPreset; this means POST is needed to create one.
func (podPresetStrategy) AllowCreateOnUpdate() bool {
	return false
}

// ValidateUpdate is the default update validation for an end user.
func (podPresetStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	validationErrorList := validation.ValidatePodPreset(obj.(*settings.PodPreset))
	updateErrorList := validation.ValidatePodPresetUpdate(obj.(*settings.PodPreset), old.(*settings.PodPreset))
	return append(validationErrorList, updateErrorList...)
}

// AllowUnconditionalUpdate is the default update policy for Pod Preset objects.
func (podPresetStrategy) AllowUnconditionalUpdate() bool {
	return true
}
