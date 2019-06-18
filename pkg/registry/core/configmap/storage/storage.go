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

package storage

import (
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/generic"
	genericregistry "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/generic/registry"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/rest"
	api "github.com/aaron-prindle/krmapiserver/pkg/apis/core"
	"github.com/aaron-prindle/krmapiserver/pkg/printers"
	printersinternal "github.com/aaron-prindle/krmapiserver/pkg/printers/internalversion"
	printerstorage "github.com/aaron-prindle/krmapiserver/pkg/printers/storage"
	"github.com/aaron-prindle/krmapiserver/pkg/registry/core/configmap"
)

// REST implements a RESTStorage for ConfigMap
type REST struct {
	*genericregistry.Store
}

// NewREST returns a RESTStorage object that will work with ConfigMap objects.
func NewREST(optsGetter generic.RESTOptionsGetter) *REST {
	store := &genericregistry.Store{
		NewFunc:                  func() runtime.Object { return &api.ConfigMap{} },
		NewListFunc:              func() runtime.Object { return &api.ConfigMapList{} },
		DefaultQualifiedResource: api.Resource("configmaps"),

		CreateStrategy: configmap.Strategy,
		UpdateStrategy: configmap.Strategy,
		DeleteStrategy: configmap.Strategy,

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
	return []string{"cm"}
}
