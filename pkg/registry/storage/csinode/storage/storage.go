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

package storage

import (
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/generic"
	genericregistry "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/generic/registry"
	storageapi "github.com/aaron-prindle/krmapiserver/pkg/apis/storage"
	"github.com/aaron-prindle/krmapiserver/pkg/registry/storage/csinode"
)

// CSINodeStorage includes storage for CSINodes and all subresources
type CSINodeStorage struct {
	CSINode *REST
}

// REST object that will work for CSINodes
type REST struct {
	*genericregistry.Store
}

// NewStorage returns a RESTStorage object that will work against CSINodes
func NewStorage(optsGetter generic.RESTOptionsGetter) *CSINodeStorage {
	store := &genericregistry.Store{
		NewFunc:                  func() runtime.Object { return &storageapi.CSINode{} },
		NewListFunc:              func() runtime.Object { return &storageapi.CSINodeList{} },
		DefaultQualifiedResource: storageapi.Resource("csinodes"),

		CreateStrategy:      csinode.Strategy,
		UpdateStrategy:      csinode.Strategy,
		DeleteStrategy:      csinode.Strategy,
		ReturnDeletedObject: true,
	}
	options := &generic.StoreOptions{RESTOptions: optsGetter}
	if err := store.CompleteWithOptions(options); err != nil {
		panic(err) // TODO: Propagate error up
	}

	return &CSINodeStorage{
		CSINode: &REST{store},
	}
}
