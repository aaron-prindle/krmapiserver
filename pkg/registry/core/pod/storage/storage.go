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
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/api/errors"
	metav1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/generic"
	genericregistry "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/generic/registry"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/registry/rest"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/storage"
	storeerr "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/storage/errors"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/util/dryrun"
	policyclient "github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/kubernetes/typed/policy/v1beta1"
	podutil "github.com/aaron-prindle/krmapiserver/pkg/api/pod"
	api "github.com/aaron-prindle/krmapiserver/pkg/apis/core"
	"github.com/aaron-prindle/krmapiserver/pkg/apis/core/validation"
	"github.com/aaron-prindle/krmapiserver/pkg/kubelet/client"
	"github.com/aaron-prindle/krmapiserver/pkg/printers"
	printersinternal "github.com/aaron-prindle/krmapiserver/pkg/printers/internalversion"
	printerstorage "github.com/aaron-prindle/krmapiserver/pkg/printers/storage"
	"github.com/aaron-prindle/krmapiserver/pkg/registry/core/pod"
	podrest "github.com/aaron-prindle/krmapiserver/pkg/registry/core/pod/rest"
)

// PodStorage includes storage for pods and all sub resources
type PodStorage struct {
	Pod         *REST
	Binding     *BindingREST
	Eviction    *EvictionREST
	Status      *StatusREST
	Log         *podrest.LogREST
	Proxy       *podrest.ProxyREST
	Exec        *podrest.ExecREST
	Attach      *podrest.AttachREST
	PortForward *podrest.PortForwardREST
}

// REST implements a RESTStorage for pods
type REST struct {
	*genericregistry.Store
	proxyTransport http.RoundTripper
}

// NewStorage returns a RESTStorage object that will work against pods.
func NewStorage(optsGetter generic.RESTOptionsGetter, k client.ConnectionInfoGetter, proxyTransport http.RoundTripper, podDisruptionBudgetClient policyclient.PodDisruptionBudgetsGetter) PodStorage {

	store := &genericregistry.Store{
		NewFunc:                  func() runtime.Object { return &api.Pod{} },
		NewListFunc:              func() runtime.Object { return &api.PodList{} },
		PredicateFunc:            pod.MatchPod,
		DefaultQualifiedResource: api.Resource("pods"),

		CreateStrategy:      pod.Strategy,
		UpdateStrategy:      pod.Strategy,
		DeleteStrategy:      pod.Strategy,
		ReturnDeletedObject: true,

		TableConvertor: printerstorage.TableConvertor{TableGenerator: printers.NewTableGenerator().With(printersinternal.AddHandlers)},
	}
	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: pod.GetAttrs, TriggerFunc: pod.NodeNameTriggerFunc}
	if err := store.CompleteWithOptions(options); err != nil {
		panic(err) // TODO: Propagate error up
	}

	statusStore := *store
	statusStore.UpdateStrategy = pod.StatusStrategy

	return PodStorage{
		Pod:         &REST{store, proxyTransport},
		Binding:     &BindingREST{store: store},
		Eviction:    newEvictionStorage(store, podDisruptionBudgetClient),
		Status:      &StatusREST{store: &statusStore},
		Log:         &podrest.LogREST{Store: store, KubeletConn: k},
		Proxy:       &podrest.ProxyREST{Store: store, ProxyTransport: proxyTransport},
		Exec:        &podrest.ExecREST{Store: store, KubeletConn: k},
		Attach:      &podrest.AttachREST{Store: store, KubeletConn: k},
		PortForward: &podrest.PortForwardREST{Store: store, KubeletConn: k},
	}
}

// Implement Redirector.
var _ = rest.Redirector(&REST{})

// ResourceLocation returns a pods location from its HostIP
func (r *REST) ResourceLocation(ctx context.Context, name string) (*url.URL, http.RoundTripper, error) {
	return pod.ResourceLocation(r, r.proxyTransport, ctx, name)
}

// Implement ShortNamesProvider
var _ rest.ShortNamesProvider = &REST{}

// ShortNames implements the ShortNamesProvider interface. Returns a list of short names for a resource.
func (r *REST) ShortNames() []string {
	return []string{"po"}
}

// Implement CategoriesProvider
var _ rest.CategoriesProvider = &REST{}

// Categories implements the CategoriesProvider interface. Returns a list of categories a resource is part of.
func (r *REST) Categories() []string {
	return []string{"all"}
}

// BindingREST implements the REST endpoint for binding pods to nodes when etcd is in use.
type BindingREST struct {
	store *genericregistry.Store
}

// NamespaceScoped fulfill rest.Scoper
func (r *BindingREST) NamespaceScoped() bool {
	return r.store.NamespaceScoped()
}

// New creates a new binding resource
func (r *BindingREST) New() runtime.Object {
	return &api.Binding{}
}

var _ = rest.Creater(&BindingREST{})

// Create ensures a pod is bound to a specific host.
func (r *BindingREST) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *metav1.CreateOptions) (out runtime.Object, err error) {
	binding := obj.(*api.Binding)

	// TODO: move me to a binding strategy
	if errs := validation.ValidatePodBinding(binding); len(errs) != 0 {
		return nil, errs.ToAggregate()
	}

	if createValidation != nil {
		if err := createValidation(binding.DeepCopyObject()); err != nil {
			return nil, err
		}
	}

	err = r.assignPod(ctx, binding.Name, binding.Target.Name, binding.Annotations, dryrun.IsDryRun(options.DryRun))
	out = &metav1.Status{Status: metav1.StatusSuccess}
	return
}

// setPodHostAndAnnotations sets the given pod's host to 'machine' if and only if it was
// previously 'oldMachine' and merges the provided annotations with those of the pod.
// Returns the current state of the pod, or an error.
func (r *BindingREST) setPodHostAndAnnotations(ctx context.Context, podID, oldMachine, machine string, annotations map[string]string, dryRun bool) (finalPod *api.Pod, err error) {
	podKey, err := r.store.KeyFunc(ctx, podID)
	if err != nil {
		return nil, err
	}
	err = r.store.Storage.GuaranteedUpdate(ctx, podKey, &api.Pod{}, false, nil, storage.SimpleUpdate(func(obj runtime.Object) (runtime.Object, error) {
		pod, ok := obj.(*api.Pod)
		if !ok {
			return nil, fmt.Errorf("unexpected object: %#v", obj)
		}
		if pod.DeletionTimestamp != nil {
			return nil, fmt.Errorf("pod %s is being deleted, cannot be assigned to a host", pod.Name)
		}
		if pod.Spec.NodeName != oldMachine {
			return nil, fmt.Errorf("pod %v is already assigned to node %q", pod.Name, pod.Spec.NodeName)
		}
		pod.Spec.NodeName = machine
		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}
		for k, v := range annotations {
			pod.Annotations[k] = v
		}
		podutil.UpdatePodCondition(&pod.Status, &api.PodCondition{
			Type:   api.PodScheduled,
			Status: api.ConditionTrue,
		})
		finalPod = pod
		return pod, nil
	}), dryRun)
	return finalPod, err
}

// assignPod assigns the given pod to the given machine.
func (r *BindingREST) assignPod(ctx context.Context, podID string, machine string, annotations map[string]string, dryRun bool) (err error) {
	if _, err = r.setPodHostAndAnnotations(ctx, podID, "", machine, annotations, dryRun); err != nil {
		err = storeerr.InterpretGetError(err, api.Resource("pods"), podID)
		err = storeerr.InterpretUpdateError(err, api.Resource("pods"), podID)
		if _, ok := err.(*errors.StatusError); !ok {
			err = errors.NewConflict(api.Resource("pods/binding"), podID, err)
		}
	}
	return
}

// StatusREST implements the REST endpoint for changing the status of a pod.
type StatusREST struct {
	store *genericregistry.Store
}

// New creates a new pod resource
func (r *StatusREST) New() runtime.Object {
	return &api.Pod{}
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
