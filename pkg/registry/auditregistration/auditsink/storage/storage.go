/*
Copyright 2018 The Kubernetes Authors.

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
	"github.com/aaron-prindle/krmapiserver/pkg/apis/auditregistration"
	auditstrategy "github.com/aaron-prindle/krmapiserver/pkg/registry/auditregistration/auditsink"
)

// REST implements a RESTStorage for audit sink against etcd
type REST struct {
	*genericregistry.Store
}

// NewREST returns a RESTStorage object that will work against audit sinks
func NewREST(optsGetter generic.RESTOptionsGetter) *REST {
	store := &genericregistry.Store{
		NewFunc:     func() runtime.Object { return &auditregistration.AuditSink{} },
		NewListFunc: func() runtime.Object { return &auditregistration.AuditSinkList{} },
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return obj.(*auditregistration.AuditSink).Name, nil
		},
		DefaultQualifiedResource: auditregistration.Resource("auditsinks"),

		CreateStrategy: auditstrategy.Strategy,
		UpdateStrategy: auditstrategy.Strategy,
		DeleteStrategy: auditstrategy.Strategy,
	}
	options := &generic.StoreOptions{RESTOptions: optsGetter}
	if err := store.CompleteWithOptions(options); err != nil {
		panic(err) // TODO: Propagate error up
	}
	return &REST{store}
}
