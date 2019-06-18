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

// Code generated by informer-gen. DO NOT EDIT.

package v1beta1

import (
	time "time"

	policyv1beta1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/policy/v1beta1"
	v1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	watch "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/watch"
	internalinterfaces "github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/informers/internalinterfaces"
	kubernetes "github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/kubernetes"
	v1beta1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/listers/policy/v1beta1"
	cache "github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/tools/cache"
)

// PodDisruptionBudgetInformer provides access to a shared informer and lister for
// PodDisruptionBudgets.
type PodDisruptionBudgetInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1beta1.PodDisruptionBudgetLister
}

type podDisruptionBudgetInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewPodDisruptionBudgetInformer constructs a new informer for PodDisruptionBudget type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewPodDisruptionBudgetInformer(client kubernetes.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredPodDisruptionBudgetInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredPodDisruptionBudgetInformer constructs a new informer for PodDisruptionBudget type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredPodDisruptionBudgetInformer(client kubernetes.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.PolicyV1beta1().PodDisruptionBudgets(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.PolicyV1beta1().PodDisruptionBudgets(namespace).Watch(options)
			},
		},
		&policyv1beta1.PodDisruptionBudget{},
		resyncPeriod,
		indexers,
	)
}

func (f *podDisruptionBudgetInformer) defaultInformer(client kubernetes.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredPodDisruptionBudgetInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *podDisruptionBudgetInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&policyv1beta1.PodDisruptionBudget{}, f.defaultInformer)
}

func (f *podDisruptionBudgetInformer) Lister() v1beta1.PodDisruptionBudgetLister {
	return v1beta1.NewPodDisruptionBudgetLister(f.Informer().GetIndexer())
}
