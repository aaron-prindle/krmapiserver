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

package storage

import (
	"time"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/authentication/authenticator"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/generic"
	genericregistry "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/generic/registry"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/rest"
	api "github.com/aaron-prindle/krmapiserver/pkg/apis/core"
	"github.com/aaron-prindle/krmapiserver/pkg/printers"
	printersinternal "github.com/aaron-prindle/krmapiserver/pkg/printers/internalversion"
	printerstorage "github.com/aaron-prindle/krmapiserver/pkg/printers/storage"
	"github.com/aaron-prindle/krmapiserver/pkg/registry/core/serviceaccount"
	token "github.com/aaron-prindle/krmapiserver/pkg/serviceaccount"
)

type REST struct {
	*genericregistry.Store
	Token *TokenREST
}

// NewREST returns a RESTStorage object that will work against service accounts.
func NewREST(optsGetter generic.RESTOptionsGetter, issuer token.TokenGenerator, auds authenticator.Audiences, max time.Duration, podStorage, secretStorage *genericregistry.Store) *REST {
	store := &genericregistry.Store{
		NewFunc:                  func() runtime.Object { return &api.ServiceAccount{} },
		NewListFunc:              func() runtime.Object { return &api.ServiceAccountList{} },
		DefaultQualifiedResource: api.Resource("serviceaccounts"),

		CreateStrategy:      serviceaccount.Strategy,
		UpdateStrategy:      serviceaccount.Strategy,
		DeleteStrategy:      serviceaccount.Strategy,
		ReturnDeletedObject: true,

		TableConvertor: printerstorage.TableConvertor{TableGenerator: printers.NewTableGenerator().With(printersinternal.AddHandlers)},
	}
	options := &generic.StoreOptions{RESTOptions: optsGetter}
	if err := store.CompleteWithOptions(options); err != nil {
		panic(err) // TODO: Propagate error up
	}

	var trest *TokenREST
	if issuer != nil && podStorage != nil && secretStorage != nil {
		trest = &TokenREST{
			svcaccts:             store,
			pods:                 podStorage,
			secrets:              secretStorage,
			issuer:               issuer,
			auds:                 auds,
			maxExpirationSeconds: int64(max.Seconds()),
		}
	}

	return &REST{
		Store: store,
		Token: trest,
	}
}

// Implement ShortNamesProvider
var _ rest.ShortNamesProvider = &REST{}

// ShortNames implements the ShortNamesProvider interface. Returns a list of short names for a resource.
func (r *REST) ShortNames() []string {
	return []string{"sa"}
}
