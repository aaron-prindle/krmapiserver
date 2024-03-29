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

package serviceaccount

import (
	"context"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/validation/field"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/storage/names"
	"github.com/aaron-prindle/krmapiserver/pkg/api/legacyscheme"
	api "github.com/aaron-prindle/krmapiserver/pkg/apis/core"
	"github.com/aaron-prindle/krmapiserver/pkg/apis/core/validation"
)

// strategy implements behavior for ServiceAccount objects
type strategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating ServiceAccount
// objects via the REST API.
var Strategy = strategy{legacyscheme.Scheme, names.SimpleNameGenerator}

func (strategy) NamespaceScoped() bool {
	return true
}

func (strategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	cleanSecretReferences(obj.(*api.ServiceAccount))
}

func (strategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return validation.ValidateServiceAccount(obj.(*api.ServiceAccount))
}

// Canonicalize normalizes the object after validation.
func (strategy) Canonicalize(obj runtime.Object) {
}

func (strategy) AllowCreateOnUpdate() bool {
	return false
}

func (strategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	cleanSecretReferences(obj.(*api.ServiceAccount))
}

func cleanSecretReferences(serviceAccount *api.ServiceAccount) {
	for i, secret := range serviceAccount.Secrets {
		serviceAccount.Secrets[i] = api.ObjectReference{Name: secret.Name}
	}
}

func (strategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateServiceAccountUpdate(obj.(*api.ServiceAccount), old.(*api.ServiceAccount))
}

func (strategy) AllowUnconditionalUpdate() bool {
	return true
}
