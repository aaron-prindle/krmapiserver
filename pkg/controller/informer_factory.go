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

package controller

import (
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime/schema"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/dynamic/dynamicinformer"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/informers"
)

// InformerFactory creates informers for each group version resource.
type InformerFactory interface {
	ForResource(resource schema.GroupVersionResource) (informers.GenericInformer, error)
	Start(stopCh <-chan struct{})
}

type informerFactory struct {
	typedInformerFactory   informers.SharedInformerFactory
	dynamicInformerFactory dynamicinformer.DynamicSharedInformerFactory
}

func (i *informerFactory) ForResource(resource schema.GroupVersionResource) (informers.GenericInformer, error) {
	informer, err := i.typedInformerFactory.ForResource(resource)
	if err != nil {
		return i.dynamicInformerFactory.ForResource(resource), nil
	}
	return informer, nil
}

func (i *informerFactory) Start(stopCh <-chan struct{}) {
	i.typedInformerFactory.Start(stopCh)
	i.dynamicInformerFactory.Start(stopCh)
}

// NewInformerFactory creates a new InformerFactory which works with both typed
// resources and dynamic resources
func NewInformerFactory(typedInformerFactory informers.SharedInformerFactory, dynamicInformerFactory dynamicinformer.DynamicSharedInformerFactory) InformerFactory {
	return &informerFactory{
		typedInformerFactory:   typedInformerFactory,
		dynamicInformerFactory: dynamicInformerFactory,
	}
}
