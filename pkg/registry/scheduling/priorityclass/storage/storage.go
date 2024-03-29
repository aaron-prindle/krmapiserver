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

package storage

import (
	"context"
	"errors"

	apierrors "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/api/errors"
	metav1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/generic"
	genericregistry "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/generic/registry"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/rest"
	"github.com/aaron-prindle/krmapiserver/pkg/apis/scheduling"
	"github.com/aaron-prindle/krmapiserver/pkg/printers"
	printersinternal "github.com/aaron-prindle/krmapiserver/pkg/printers/internalversion"
	printerstorage "github.com/aaron-prindle/krmapiserver/pkg/printers/storage"
	"github.com/aaron-prindle/krmapiserver/pkg/registry/scheduling/priorityclass"
)

// REST implements a RESTStorage for priority classes against etcd
type REST struct {
	*genericregistry.Store
}

// NewREST returns a RESTStorage object that will work against priority classes.
func NewREST(optsGetter generic.RESTOptionsGetter) *REST {
	store := &genericregistry.Store{
		NewFunc:                  func() runtime.Object { return &scheduling.PriorityClass{} },
		NewListFunc:              func() runtime.Object { return &scheduling.PriorityClassList{} },
		DefaultQualifiedResource: scheduling.Resource("priorityclasses"),

		CreateStrategy: priorityclass.Strategy,
		UpdateStrategy: priorityclass.Strategy,
		DeleteStrategy: priorityclass.Strategy,

		TableConvertor: printerstorage.TableConvertor{TableGenerator: printers.NewTableGenerator().With(printersinternal.AddHandlers)},
	}
	options := &generic.StoreOptions{RESTOptions: optsGetter}
	if err := store.CompleteWithOptions(options); err != nil {
		panic(err) // TODO: Propagate error up
	}

	return &REST{store}
}

// Implement ShortNamesProvider
var _ rest.ShortNamesProvider = &REST{}

// ShortNames implements the ShortNamesProvider interface. Returns a list of short names for a resource.
func (r *REST) ShortNames() []string {
	return []string{"pc"}
}

// Delete ensures that system priority classes are not deleted.
func (r *REST) Delete(ctx context.Context, name string, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions) (runtime.Object, bool, error) {
	for _, spc := range scheduling.SystemPriorityClasses() {
		if name == spc.Name {
			return nil, false, apierrors.NewForbidden(scheduling.Resource("priorityclasses"), spc.Name, errors.New("this is a system priority class and cannot be deleted"))
		}
	}

	return r.Store.Delete(ctx, name, deleteValidation, options)
}
