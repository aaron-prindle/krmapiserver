/*
Copyright 2015 The Kubernetes Authors.

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

package secret

import (
	"context"
	"fmt"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/api/errors"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/fields"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/labels"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/validation/field"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/generic"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/rest"
	pkgstorage "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/storage"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/storage/names"
	"github.com/aaron-prindle/krmapiserver/pkg/api/legacyscheme"
	api "github.com/aaron-prindle/krmapiserver/pkg/apis/core"
	"github.com/aaron-prindle/krmapiserver/pkg/apis/core/validation"
)

// strategy implements behavior for Secret objects
type strategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating Secret
// objects via the REST API.
var Strategy = strategy{legacyscheme.Scheme, names.SimpleNameGenerator}

var _ = rest.RESTCreateStrategy(Strategy)

var _ = rest.RESTUpdateStrategy(Strategy)

func (strategy) NamespaceScoped() bool {
	return true
}

func (strategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (strategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return validation.ValidateSecret(obj.(*api.Secret))
}

func (strategy) Canonicalize(obj runtime.Object) {
}

func (strategy) AllowCreateOnUpdate() bool {
	return false
}

func (strategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (strategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateSecretUpdate(obj.(*api.Secret), old.(*api.Secret))
}

func (strategy) AllowUnconditionalUpdate() bool {
	return true
}

func (s strategy) Export(ctx context.Context, obj runtime.Object, exact bool) error {
	t, ok := obj.(*api.Secret)
	if !ok {
		// unexpected programmer error
		return fmt.Errorf("unexpected object: %v", obj)
	}
	s.PrepareForCreate(ctx, obj)
	if exact {
		return nil
	}
	// secrets that are tied to the UID of a service account cannot be exported anyway
	if t.Type == api.SecretTypeServiceAccountToken || len(t.Annotations[api.ServiceAccountUIDKey]) > 0 {
		errs := []*field.Error{
			field.Invalid(field.NewPath("type"), t, "can not export service account secrets"),
		}
		return errors.NewInvalid(api.Kind("Secret"), t.Name, errs)
	}
	return nil
}

// GetAttrs returns labels and fields of a given object for filtering purposes.
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	secret, ok := obj.(*api.Secret)
	if !ok {
		return nil, nil, fmt.Errorf("not a secret")
	}
	return labels.Set(secret.Labels), SelectableFields(secret), nil
}

// Matcher returns a generic matcher for a given label and field selector.
func Matcher(label labels.Selector, field fields.Selector) pkgstorage.SelectionPredicate {
	return pkgstorage.SelectionPredicate{
		Label:       label,
		Field:       field,
		GetAttrs:    GetAttrs,
		IndexFields: []string{"metadata.name"},
	}
}

func SecretNameTriggerFunc(obj runtime.Object) []pkgstorage.MatchValue {
	secret := obj.(*api.Secret)
	result := pkgstorage.MatchValue{IndexName: "metadata.name", Value: secret.ObjectMeta.Name}
	return []pkgstorage.MatchValue{result}
}

// SelectableFields returns a field set that can be used for filter selection
func SelectableFields(obj *api.Secret) fields.Set {
	objectMetaFieldsSet := generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
	secretSpecificFieldsSet := fields.Set{
		"type": string(obj.Type),
	}
	return generic.MergeFieldsSets(objectMetaFieldsSet, secretSpecificFieldsSet)
}
