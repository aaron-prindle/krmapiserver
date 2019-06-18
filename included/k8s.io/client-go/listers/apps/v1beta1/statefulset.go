/*
Copyright The Kubernetes Authors.

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

// Code generated by lister-gen. DO NOT EDIT.

package v1beta1

import (
	v1beta1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/apps/v1beta1"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/api/errors"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/labels"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/tools/cache"
)

// StatefulSetLister helps list StatefulSets.
type StatefulSetLister interface {
	// List lists all StatefulSets in the indexer.
	List(selector labels.Selector) (ret []*v1beta1.StatefulSet, err error)
	// StatefulSets returns an object that can list and get StatefulSets.
	StatefulSets(namespace string) StatefulSetNamespaceLister
	StatefulSetListerExpansion
}

// statefulSetLister implements the StatefulSetLister interface.
type statefulSetLister struct {
	indexer cache.Indexer
}

// NewStatefulSetLister returns a new StatefulSetLister.
func NewStatefulSetLister(indexer cache.Indexer) StatefulSetLister {
	return &statefulSetLister{indexer: indexer}
}

// List lists all StatefulSets in the indexer.
func (s *statefulSetLister) List(selector labels.Selector) (ret []*v1beta1.StatefulSet, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.StatefulSet))
	})
	return ret, err
}

// StatefulSets returns an object that can list and get StatefulSets.
func (s *statefulSetLister) StatefulSets(namespace string) StatefulSetNamespaceLister {
	return statefulSetNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// StatefulSetNamespaceLister helps list and get StatefulSets.
type StatefulSetNamespaceLister interface {
	// List lists all StatefulSets in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1beta1.StatefulSet, err error)
	// Get retrieves the StatefulSet from the indexer for a given namespace and name.
	Get(name string) (*v1beta1.StatefulSet, error)
	StatefulSetNamespaceListerExpansion
}

// statefulSetNamespaceLister implements the StatefulSetNamespaceLister
// interface.
type statefulSetNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all StatefulSets in the indexer for a given namespace.
func (s statefulSetNamespaceLister) List(selector labels.Selector) (ret []*v1beta1.StatefulSet, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.StatefulSet))
	})
	return ret, err
}

// Get retrieves the StatefulSet from the indexer for a given namespace and name.
func (s statefulSetNamespaceLister) Get(name string) (*v1beta1.StatefulSet, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1beta1.Resource("statefulset"), name)
	}
	return obj.(*v1beta1.StatefulSet), nil
}
