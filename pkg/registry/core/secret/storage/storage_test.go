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
	"testing"

	metav1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/fields"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/labels"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/generic"
	genericregistrytest "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/generic/testing"
	etcdtesting "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/storage/etcd/testing"
	api "github.com/aaron-prindle/krmapiserver/pkg/apis/core"
	"github.com/aaron-prindle/krmapiserver/pkg/registry/registrytest"
)

func newStorage(t *testing.T) (*REST, *etcdtesting.EtcdTestServer) {
	etcdStorage, server := registrytest.NewEtcdStorage(t, "")
	restOptions := generic.RESTOptions{
		StorageConfig:           etcdStorage,
		Decorator:               generic.UndecoratedStorage,
		DeleteCollectionWorkers: 1,
		ResourcePrefix:          "secrets",
	}
	return NewREST(restOptions), server
}

func validNewSecret(name string) *api.Secret {
	return &api.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Data: map[string][]byte{
			"test": []byte("data"),
		},
	}
}

func TestCreate(t *testing.T) {
	storage, server := newStorage(t)
	defer server.Terminate(t)
	defer storage.Store.DestroyFunc()
	test := genericregistrytest.New(t, storage.Store)
	secret := validNewSecret("foo")
	secret.ObjectMeta = metav1.ObjectMeta{GenerateName: "foo-"}
	test.TestCreate(
		// valid
		secret,
		// invalid
		&api.Secret{},
		&api.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "name"},
			Data:       map[string][]byte{"name with spaces": []byte("")},
		},
		&api.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "name"},
			Data:       map[string][]byte{"~.dotfile": []byte("")},
		},
	)
}

func TestUpdate(t *testing.T) {
	storage, server := newStorage(t)
	defer server.Terminate(t)
	defer storage.Store.DestroyFunc()
	test := genericregistrytest.New(t, storage.Store)
	test.TestUpdate(
		// valid
		validNewSecret("foo"),
		// updateFunc
		func(obj runtime.Object) runtime.Object {
			object := obj.(*api.Secret)
			object.Data["othertest"] = []byte("otherdata")
			return object
		},
	)
}

func TestDelete(t *testing.T) {
	storage, server := newStorage(t)
	defer server.Terminate(t)
	defer storage.Store.DestroyFunc()
	test := genericregistrytest.New(t, storage.Store)
	test.TestDelete(validNewSecret("foo"))
}

func TestGet(t *testing.T) {
	storage, server := newStorage(t)
	defer server.Terminate(t)
	defer storage.Store.DestroyFunc()
	test := genericregistrytest.New(t, storage.Store)
	test.TestGet(validNewSecret("foo"))
}

func TestList(t *testing.T) {
	storage, server := newStorage(t)
	defer server.Terminate(t)
	defer storage.Store.DestroyFunc()
	test := genericregistrytest.New(t, storage.Store)
	test.TestList(validNewSecret("foo"))
}

func TestWatch(t *testing.T) {
	storage, server := newStorage(t)
	defer server.Terminate(t)
	defer storage.Store.DestroyFunc()
	test := genericregistrytest.New(t, storage.Store)
	test.TestWatch(
		validNewSecret("foo"),
		// matching labels
		[]labels.Set{},
		// not matching labels
		[]labels.Set{
			{"foo": "bar"},
		},
		// matching fields
		[]fields.Set{},
		// not matching fields
		[]fields.Set{
			{"metadata.name": "bar"},
			{"name": "foo"},
		},
	)
}
