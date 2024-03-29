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
	"context"
	"fmt"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/api/errors"
	metav1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime/schema"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/generic"
	genericregistry "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/generic/registry"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/rest"
	"github.com/aaron-prindle/krmapiserver/pkg/apis/apps"
	appsv1beta1 "github.com/aaron-prindle/krmapiserver/pkg/apis/apps/v1beta1"
	appsv1beta2 "github.com/aaron-prindle/krmapiserver/pkg/apis/apps/v1beta2"
	"github.com/aaron-prindle/krmapiserver/pkg/apis/autoscaling"
	autoscalingv1 "github.com/aaron-prindle/krmapiserver/pkg/apis/autoscaling/v1"
	autoscalingvalidation "github.com/aaron-prindle/krmapiserver/pkg/apis/autoscaling/validation"
	"github.com/aaron-prindle/krmapiserver/pkg/printers"
	printersinternal "github.com/aaron-prindle/krmapiserver/pkg/printers/internalversion"
	printerstorage "github.com/aaron-prindle/krmapiserver/pkg/printers/storage"
	"github.com/aaron-prindle/krmapiserver/pkg/registry/apps/statefulset"
)

// StatefulSetStorage includes dummy storage for StatefulSets, and their Status and Scale subresource.
type StatefulSetStorage struct {
	StatefulSet *REST
	Status      *StatusREST
	Scale       *ScaleREST
}

func NewStorage(optsGetter generic.RESTOptionsGetter) StatefulSetStorage {
	statefulSetRest, statefulSetStatusRest := NewREST(optsGetter)

	return StatefulSetStorage{
		StatefulSet: statefulSetRest,
		Status:      statefulSetStatusRest,
		Scale:       &ScaleREST{store: statefulSetRest.Store},
	}
}

// rest implements a RESTStorage for statefulsets against etcd
type REST struct {
	*genericregistry.Store
}

// NewREST returns a RESTStorage object that will work against statefulsets.
func NewREST(optsGetter generic.RESTOptionsGetter) (*REST, *StatusREST) {
	store := &genericregistry.Store{
		NewFunc:                  func() runtime.Object { return &apps.StatefulSet{} },
		NewListFunc:              func() runtime.Object { return &apps.StatefulSetList{} },
		DefaultQualifiedResource: apps.Resource("statefulsets"),

		CreateStrategy: statefulset.Strategy,
		UpdateStrategy: statefulset.Strategy,
		DeleteStrategy: statefulset.Strategy,

		TableConvertor: printerstorage.TableConvertor{TableGenerator: printers.NewTableGenerator().With(printersinternal.AddHandlers)},
	}
	options := &generic.StoreOptions{RESTOptions: optsGetter}
	if err := store.CompleteWithOptions(options); err != nil {
		panic(err) // TODO: Propagate error up
	}

	statusStore := *store
	statusStore.UpdateStrategy = statefulset.StatusStrategy
	return &REST{store}, &StatusREST{store: &statusStore}
}

// Implement CategoriesProvider
var _ rest.CategoriesProvider = &REST{}

// Categories implements the CategoriesProvider interface. Returns a list of categories a resource is part of.
func (r *REST) Categories() []string {
	return []string{"all"}
}

// StatusREST implements the REST endpoint for changing the status of an statefulSet
type StatusREST struct {
	store *genericregistry.Store
}

func (r *StatusREST) New() runtime.Object {
	return &apps.StatefulSet{}
}

// Get retrieves the object from the storage. It is required to support Patch.
func (r *StatusREST) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return r.store.Get(ctx, name, options)
}

// Update alters the status subset of an object.
func (r *StatusREST) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {
	// We are explicitly setting forceAllowCreate to false in the call to the underlying storage because
	// subresources should never allow create on update.
	return r.store.Update(ctx, name, objInfo, createValidation, updateValidation, false, options)
}

// Implement ShortNamesProvider
var _ rest.ShortNamesProvider = &REST{}

// ShortNames implements the ShortNamesProvider interface. Returns a list of short names for a resource.
func (r *REST) ShortNames() []string {
	return []string{"sts"}
}

type ScaleREST struct {
	store *genericregistry.Store
}

// ScaleREST implements Patcher
var _ = rest.Patcher(&ScaleREST{})
var _ = rest.GroupVersionKindProvider(&ScaleREST{})

func (r *ScaleREST) GroupVersionKind(containingGV schema.GroupVersion) schema.GroupVersionKind {
	switch containingGV {
	case appsv1beta1.SchemeGroupVersion:
		return appsv1beta1.SchemeGroupVersion.WithKind("Scale")
	case appsv1beta2.SchemeGroupVersion:
		return appsv1beta2.SchemeGroupVersion.WithKind("Scale")
	default:
		return autoscalingv1.SchemeGroupVersion.WithKind("Scale")
	}
}

// New creates a new Scale object
func (r *ScaleREST) New() runtime.Object {
	return &autoscaling.Scale{}
}

func (r *ScaleREST) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	obj, err := r.store.Get(ctx, name, options)
	if err != nil {
		return nil, err
	}
	ss := obj.(*apps.StatefulSet)
	scale, err := scaleFromStatefulSet(ss)
	if err != nil {
		return nil, errors.NewBadRequest(fmt.Sprintf("%v", err))
	}
	return scale, err
}

func (r *ScaleREST) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {
	obj, err := r.store.Get(ctx, name, &metav1.GetOptions{})
	if err != nil {
		return nil, false, err
	}
	ss := obj.(*apps.StatefulSet)

	oldScale, err := scaleFromStatefulSet(ss)
	if err != nil {
		return nil, false, err
	}

	obj, err = objInfo.UpdatedObject(ctx, oldScale)
	if err != nil {
		return nil, false, err
	}
	if obj == nil {
		return nil, false, errors.NewBadRequest(fmt.Sprintf("nil update passed to Scale"))
	}
	scale, ok := obj.(*autoscaling.Scale)
	if !ok {
		return nil, false, errors.NewBadRequest(fmt.Sprintf("wrong object passed to Scale update: %v", obj))
	}

	if errs := autoscalingvalidation.ValidateScale(scale); len(errs) > 0 {
		return nil, false, errors.NewInvalid(autoscaling.Kind("Scale"), scale.Name, errs)
	}

	ss.Spec.Replicas = scale.Spec.Replicas
	ss.ResourceVersion = scale.ResourceVersion
	obj, _, err = r.store.Update(
		ctx,
		ss.Name,
		rest.DefaultUpdatedObjectInfo(ss),
		toScaleCreateValidation(createValidation),
		toScaleUpdateValidation(updateValidation),
		false,
		options,
	)
	if err != nil {
		return nil, false, err
	}
	ss = obj.(*apps.StatefulSet)
	newScale, err := scaleFromStatefulSet(ss)
	if err != nil {
		return nil, false, errors.NewBadRequest(fmt.Sprintf("%v", err))
	}
	return newScale, false, err
}

func toScaleCreateValidation(f rest.ValidateObjectFunc) rest.ValidateObjectFunc {
	return func(obj runtime.Object) error {
		scale, err := scaleFromStatefulSet(obj.(*apps.StatefulSet))
		if err != nil {
			return err
		}
		return f(scale)
	}
}

func toScaleUpdateValidation(f rest.ValidateObjectUpdateFunc) rest.ValidateObjectUpdateFunc {
	return func(obj, old runtime.Object) error {
		newScale, err := scaleFromStatefulSet(obj.(*apps.StatefulSet))
		if err != nil {
			return err
		}
		oldScale, err := scaleFromStatefulSet(old.(*apps.StatefulSet))
		if err != nil {
			return err
		}
		return f(newScale, oldScale)
	}
}

// scaleFromStatefulSet returns a scale subresource for a statefulset.
func scaleFromStatefulSet(ss *apps.StatefulSet) (*autoscaling.Scale, error) {
	selector, err := metav1.LabelSelectorAsSelector(ss.Spec.Selector)
	if err != nil {
		return nil, err
	}
	return &autoscaling.Scale{
		// TODO: Create a variant of ObjectMeta type that only contains the fields below.
		ObjectMeta: metav1.ObjectMeta{
			Name:              ss.Name,
			Namespace:         ss.Namespace,
			UID:               ss.UID,
			ResourceVersion:   ss.ResourceVersion,
			CreationTimestamp: ss.CreationTimestamp,
		},
		Spec: autoscaling.ScaleSpec{
			Replicas: ss.Spec.Replicas,
		},
		Status: autoscaling.ScaleStatus{
			Replicas: ss.Status.Replicas,
			Selector: selector.String(),
		},
	}, nil
}
