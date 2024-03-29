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

package podtemplate

import (
	"context"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/validation/field"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/storage/names"
	"github.com/aaron-prindle/krmapiserver/pkg/api/legacyscheme"
	"github.com/aaron-prindle/krmapiserver/pkg/api/pod"
	api "github.com/aaron-prindle/krmapiserver/pkg/apis/core"
	corevalidation "github.com/aaron-prindle/krmapiserver/pkg/apis/core/validation"
)

// podTemplateStrategy implements behavior for PodTemplates
type podTemplateStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating PodTemplate
// objects via the REST API.
var Strategy = podTemplateStrategy{legacyscheme.Scheme, names.SimpleNameGenerator}

// NamespaceScoped is true for pod templates.
func (podTemplateStrategy) NamespaceScoped() bool {
	return true
}

// PrepareForCreate clears fields that are not allowed to be set by end users on creation.
func (podTemplateStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	template := obj.(*api.PodTemplate)

	pod.DropDisabledTemplateFields(&template.Template, nil)
}

// Validate validates a new pod template.
func (podTemplateStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	template := obj.(*api.PodTemplate)
	allErrs := corevalidation.ValidatePodTemplate(template)
	allErrs = append(allErrs, corevalidation.ValidateConditionalPodTemplate(&template.Template, nil, field.NewPath("template"))...)
	return allErrs
}

// Canonicalize normalizes the object after validation.
func (podTemplateStrategy) Canonicalize(obj runtime.Object) {
}

// AllowCreateOnUpdate is false for pod templates.
func (podTemplateStrategy) AllowCreateOnUpdate() bool {
	return false
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (podTemplateStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newTemplate := obj.(*api.PodTemplate)
	oldTemplate := old.(*api.PodTemplate)

	pod.DropDisabledTemplateFields(&newTemplate.Template, &oldTemplate.Template)
}

// ValidateUpdate is the default update validation for an end user.
func (podTemplateStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	template := obj.(*api.PodTemplate)
	oldTemplate := old.(*api.PodTemplate)
	allErrs := corevalidation.ValidatePodTemplateUpdate(template, oldTemplate)
	allErrs = append(allErrs, corevalidation.ValidateConditionalPodTemplate(&template.Template, &oldTemplate.Template, field.NewPath("template"))...)
	return allErrs
}

func (podTemplateStrategy) AllowUnconditionalUpdate() bool {
	return true
}

func (podTemplateStrategy) Export(ctx context.Context, obj runtime.Object, exact bool) error {
	// Do nothing
	return nil
}
