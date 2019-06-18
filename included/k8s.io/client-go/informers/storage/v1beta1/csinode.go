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

	storagev1beta1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/storage/v1beta1"
	v1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	watch "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/watch"
	internalinterfaces "github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/informers/internalinterfaces"
	kubernetes "github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/kubernetes"
	v1beta1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/listers/storage/v1beta1"
	cache "github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/tools/cache"
)

// CSINodeInformer provides access to a shared informer and lister for
// CSINodes.
type CSINodeInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1beta1.CSINodeLister
}

type cSINodeInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewCSINodeInformer constructs a new informer for CSINode type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewCSINodeInformer(client kubernetes.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredCSINodeInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredCSINodeInformer constructs a new informer for CSINode type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredCSINodeInformer(client kubernetes.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.StorageV1beta1().CSINodes().List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.StorageV1beta1().CSINodes().Watch(options)
			},
		},
		&storagev1beta1.CSINode{},
		resyncPeriod,
		indexers,
	)
}

func (f *cSINodeInformer) defaultInformer(client kubernetes.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredCSINodeInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *cSINodeInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&storagev1beta1.CSINode{}, f.defaultInformer)
}

func (f *cSINodeInformer) Lister() v1beta1.CSINodeLister {
	return v1beta1.NewCSINodeLister(f.Informer().GetIndexer())
}
