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
	"net/http"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/api/errors"
	metav1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime/schema"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/generic"
	genericregistry "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/generic/registry"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/rest"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/storage"
	storeerr "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/storage/errors"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/util/dryrun"
	"github.com/aaron-prindle/krmapiserver/pkg/apis/apps"
	appsv1beta1 "github.com/aaron-prindle/krmapiserver/pkg/apis/apps/v1beta1"
	appsv1beta2 "github.com/aaron-prindle/krmapiserver/pkg/apis/apps/v1beta2"
	appsvalidation "github.com/aaron-prindle/krmapiserver/pkg/apis/apps/validation"
	"github.com/aaron-prindle/krmapiserver/pkg/apis/autoscaling"
	autoscalingv1 "github.com/aaron-prindle/krmapiserver/pkg/apis/autoscaling/v1"
	autoscalingvalidation "github.com/aaron-prindle/krmapiserver/pkg/apis/autoscaling/validation"
	extensionsv1beta1 "github.com/aaron-prindle/krmapiserver/pkg/apis/extensions/v1beta1"
	"github.com/aaron-prindle/krmapiserver/pkg/printers"
	printersinternal "github.com/aaron-prindle/krmapiserver/pkg/printers/internalversion"
	printerstorage "github.com/aaron-prindle/krmapiserver/pkg/printers/storage"
	"github.com/aaron-prindle/krmapiserver/pkg/registry/apps/deployment"
)

// DeploymentStorage includes dummy storage for Deployments and for Scale subresource.
type DeploymentStorage struct {
	Deployment *REST
	Status     *StatusREST
	Scale      *ScaleREST
	Rollback   *RollbackREST
}

func NewStorage(optsGetter generic.RESTOptionsGetter) DeploymentStorage {
	deploymentRest, deploymentStatusRest, deploymentRollbackRest := NewREST(optsGetter)

	return DeploymentStorage{
		Deployment: deploymentRest,
		Status:     deploymentStatusRest,
		Scale:      &ScaleREST{store: deploymentRest.Store},
		Rollback:   deploymentRollbackRest,
	}
}

type REST struct {
	*genericregistry.Store
	categories []string
}

// NewREST returns a RESTStorage object that will work against deployments.
func NewREST(optsGetter generic.RESTOptionsGetter) (*REST, *StatusREST, *RollbackREST) {
	store := &genericregistry.Store{
		NewFunc:                  func() runtime.Object { return &apps.Deployment{} },
		NewListFunc:              func() runtime.Object { return &apps.DeploymentList{} },
		DefaultQualifiedResource: apps.Resource("deployments"),

		CreateStrategy: deployment.Strategy,
		UpdateStrategy: deployment.Strategy,
		DeleteStrategy: deployment.Strategy,

		TableConvertor: printerstorage.TableConvertor{TableGenerator: printers.NewTableGenerator().With(printersinternal.AddHandlers)},
	}
	options := &generic.StoreOptions{RESTOptions: optsGetter}
	if err := store.CompleteWithOptions(options); err != nil {
		panic(err) // TODO: Propagate error up
	}

	statusStore := *store
	statusStore.UpdateStrategy = deployment.StatusStrategy
	return &REST{store, []string{"all"}}, &StatusREST{store: &statusStore}, &RollbackREST{store: store}
}

// Implement ShortNamesProvider
var _ rest.ShortNamesProvider = &REST{}

// ShortNames implements the ShortNamesProvider interface. Returns a list of short names for a resource.
func (r *REST) ShortNames() []string {
	return []string{"deploy"}
}

// Implement CategoriesProvider
var _ rest.CategoriesProvider = &REST{}

// Categories implements the CategoriesProvider interface. Returns a list of categories a resource is part of.
func (r *REST) Categories() []string {
	return r.categories
}

func (r *REST) WithCategories(categories []string) *REST {
	r.categories = categories
	return r
}

// StatusREST implements the REST endpoint for changing the status of a deployment
type StatusREST struct {
	store *genericregistry.Store
}

func (r *StatusREST) New() runtime.Object {
	return &apps.Deployment{}
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

// RollbackREST implements the REST endpoint for initiating the rollback of a deployment
type RollbackREST struct {
	store *genericregistry.Store
}

// ProducesMIMETypes returns a list of the MIME types the specified HTTP verb (GET, POST, DELETE,
// PATCH) can respond with.
func (r *RollbackREST) ProducesMIMETypes(verb string) []string {
	return nil
}

// ProducesObject returns an object the specified HTTP verb respond with. It will overwrite storage object if
// it is not nil. Only the type of the return object matters, the value will be ignored.
func (r *RollbackREST) ProducesObject(verb string) interface{} {
	return metav1.Status{}
}

var _ = rest.StorageMetadata(&RollbackREST{})

// New creates a rollback
func (r *RollbackREST) New() runtime.Object {
	return &apps.DeploymentRollback{}
}

var _ = rest.NamedCreater(&RollbackREST{})

func (r *RollbackREST) Create(ctx context.Context, name string, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	rollback, ok := obj.(*apps.DeploymentRollback)
	if !ok {
		return nil, errors.NewBadRequest(fmt.Sprintf("not a DeploymentRollback: %#v", obj))
	}

	if errs := appsvalidation.ValidateDeploymentRollback(rollback); len(errs) != 0 {
		return nil, errors.NewInvalid(apps.Kind("DeploymentRollback"), rollback.Name, errs)
	}
	if name != rollback.Name {
		return nil, errors.NewBadRequest("name in URL does not match name in DeploymentRollback object")
	}

	if createValidation != nil {
		if err := createValidation(obj.DeepCopyObject()); err != nil {
			return nil, err
		}
	}

	// Update the Deployment with information in DeploymentRollback to trigger rollback
	err := r.rollbackDeployment(ctx, rollback.Name, &rollback.RollbackTo, rollback.UpdatedAnnotations, dryrun.IsDryRun(options.DryRun))
	if err != nil {
		return nil, err
	}
	return &metav1.Status{
		Status:  metav1.StatusSuccess,
		Message: fmt.Sprintf("rollback request for deployment %q succeeded", rollback.Name),
		Code:    http.StatusOK,
	}, nil
}

func (r *RollbackREST) rollbackDeployment(ctx context.Context, deploymentID string, config *apps.RollbackConfig, annotations map[string]string, dryRun bool) error {
	if _, err := r.setDeploymentRollback(ctx, deploymentID, config, annotations, dryRun); err != nil {
		err = storeerr.InterpretGetError(err, apps.Resource("deployments"), deploymentID)
		err = storeerr.InterpretUpdateError(err, apps.Resource("deployments"), deploymentID)
		if _, ok := err.(*errors.StatusError); !ok {
			err = errors.NewInternalError(err)
		}
		return err
	}
	return nil
}

func (r *RollbackREST) setDeploymentRollback(ctx context.Context, deploymentID string, config *apps.RollbackConfig, annotations map[string]string, dryRun bool) (*apps.Deployment, error) {
	dKey, err := r.store.KeyFunc(ctx, deploymentID)
	if err != nil {
		return nil, err
	}
	var finalDeployment *apps.Deployment
	err = r.store.Storage.GuaranteedUpdate(ctx, dKey, &apps.Deployment{}, false, nil, storage.SimpleUpdate(func(obj runtime.Object) (runtime.Object, error) {
		d, ok := obj.(*apps.Deployment)
		if !ok {
			return nil, fmt.Errorf("unexpected object: %#v", obj)
		}
		if d.Annotations == nil {
			d.Annotations = make(map[string]string)
		}
		for k, v := range annotations {
			d.Annotations[k] = v
		}
		d.Spec.RollbackTo = config
		finalDeployment = d
		return d, nil
	}), dryRun)
	return finalDeployment, err
}

type ScaleREST struct {
	store *genericregistry.Store
}

// ScaleREST implements Patcher
var _ = rest.Patcher(&ScaleREST{})
var _ = rest.GroupVersionKindProvider(&ScaleREST{})

func (r *ScaleREST) GroupVersionKind(containingGV schema.GroupVersion) schema.GroupVersionKind {
	switch containingGV {
	case extensionsv1beta1.SchemeGroupVersion:
		return extensionsv1beta1.SchemeGroupVersion.WithKind("Scale")
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
		return nil, errors.NewNotFound(apps.Resource("deployments/scale"), name)
	}
	deployment := obj.(*apps.Deployment)
	scale, err := scaleFromDeployment(deployment)
	if err != nil {
		return nil, errors.NewBadRequest(fmt.Sprintf("%v", err))
	}
	return scale, nil
}

func (r *ScaleREST) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {
	obj, err := r.store.Get(ctx, name, &metav1.GetOptions{})
	if err != nil {
		return nil, false, errors.NewNotFound(apps.Resource("deployments/scale"), name)
	}
	deployment := obj.(*apps.Deployment)

	oldScale, err := scaleFromDeployment(deployment)
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
		return nil, false, errors.NewBadRequest(fmt.Sprintf("expected input object type to be Scale, but %T", obj))
	}

	if errs := autoscalingvalidation.ValidateScale(scale); len(errs) > 0 {
		return nil, false, errors.NewInvalid(autoscaling.Kind("Scale"), name, errs)
	}

	deployment.Spec.Replicas = scale.Spec.Replicas
	deployment.ResourceVersion = scale.ResourceVersion
	obj, _, err = r.store.Update(
		ctx,
		deployment.Name,
		rest.DefaultUpdatedObjectInfo(deployment),
		toScaleCreateValidation(createValidation),
		toScaleUpdateValidation(updateValidation),
		false,
		options,
	)
	if err != nil {
		return nil, false, err
	}
	deployment = obj.(*apps.Deployment)
	newScale, err := scaleFromDeployment(deployment)
	if err != nil {
		return nil, false, errors.NewBadRequest(fmt.Sprintf("%v", err))
	}
	return newScale, false, nil
}

func toScaleCreateValidation(f rest.ValidateObjectFunc) rest.ValidateObjectFunc {
	return func(obj runtime.Object) error {
		scale, err := scaleFromDeployment(obj.(*apps.Deployment))
		if err != nil {
			return err
		}
		return f(scale)
	}
}

func toScaleUpdateValidation(f rest.ValidateObjectUpdateFunc) rest.ValidateObjectUpdateFunc {
	return func(obj, old runtime.Object) error {
		newScale, err := scaleFromDeployment(obj.(*apps.Deployment))
		if err != nil {
			return err
		}
		oldScale, err := scaleFromDeployment(old.(*apps.Deployment))
		if err != nil {
			return err
		}
		return f(newScale, oldScale)
	}
}

// scaleFromDeployment returns a scale subresource for a deployment.
func scaleFromDeployment(deployment *apps.Deployment) (*autoscaling.Scale, error) {
	selector, err := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
	if err != nil {
		return nil, err
	}
	return &autoscaling.Scale{
		// TODO: Create a variant of ObjectMeta type that only contains the fields below.
		ObjectMeta: metav1.ObjectMeta{
			Name:              deployment.Name,
			Namespace:         deployment.Namespace,
			UID:               deployment.UID,
			ResourceVersion:   deployment.ResourceVersion,
			CreationTimestamp: deployment.CreationTimestamp,
		},
		Spec: autoscaling.ScaleSpec{
			Replicas: deployment.Spec.Replicas,
		},
		Status: autoscaling.ScaleStatus{
			Replicas: deployment.Status.Replicas,
			Selector: selector.String(),
		},
	}, nil
}
