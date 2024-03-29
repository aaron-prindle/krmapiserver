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

// Package mutating delegates admission checks to dynamically configured
// mutating webhooks.
package mutating

import (
	"context"
	"fmt"
	"time"

	jsonpatch "github.com/aaron-prindle/krmapiserver/included/github.com/evanphx/json-patch"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/klog"

	admissionv1beta1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/admission/v1beta1"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/api/admissionregistration/v1beta1"
	apierrors "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/api/errors"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime/serializer/json"
	utilruntime "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/admission"
	admissionmetrics "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/admission/metrics"
	webhookerrors "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/admission/plugin/webhook/errors"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/admission/plugin/webhook/generic"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/admission/plugin/webhook/request"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/admission/plugin/webhook/util"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/util/webhook"
)

type mutatingDispatcher struct {
	cm     *webhook.ClientManager
	plugin *Plugin
}

func newMutatingDispatcher(p *Plugin) func(cm *webhook.ClientManager) generic.Dispatcher {
	return func(cm *webhook.ClientManager) generic.Dispatcher {
		return &mutatingDispatcher{cm, p}
	}
}

var _ generic.Dispatcher = &mutatingDispatcher{}

func (a *mutatingDispatcher) Dispatch(ctx context.Context, attr admission.Attributes, o admission.ObjectInterfaces, relevantHooks []*generic.WebhookInvocation) error {
	var versionedAttr *generic.VersionedAttributes
	for _, invocation := range relevantHooks {
		hook := invocation.Webhook
		if versionedAttr == nil {
			// First webhook, create versioned attributes
			var err error
			if versionedAttr, err = generic.NewVersionedAttributes(attr, invocation.Kind, o); err != nil {
				return apierrors.NewInternalError(err)
			}
		} else {
			// Subsequent webhook, convert existing versioned attributes to this webhook's version
			if err := generic.ConvertVersionedAttributes(versionedAttr, invocation.Kind, o); err != nil {
				return apierrors.NewInternalError(err)
			}
		}

		t := time.Now()
		err := a.callAttrMutatingHook(ctx, invocation, versionedAttr, o)
		admissionmetrics.Metrics.ObserveWebhook(time.Since(t), err != nil, versionedAttr.Attributes, "admit", hook.Name)
		if err == nil {
			continue
		}

		ignoreClientCallFailures := hook.FailurePolicy != nil && *hook.FailurePolicy == v1beta1.Ignore
		if callErr, ok := err.(*webhook.ErrCallingWebhook); ok {
			if ignoreClientCallFailures {
				klog.Warningf("Failed calling webhook, failing open %v: %v", hook.Name, callErr)
				utilruntime.HandleError(callErr)
				continue
			}
			klog.Warningf("Failed calling webhook, failing closed %v: %v", hook.Name, err)
			return apierrors.NewInternalError(err)
		}
		return err
	}

	// convert versionedAttr.VersionedObject to the internal version in the underlying admission.Attributes
	if versionedAttr != nil && versionedAttr.VersionedObject != nil && versionedAttr.Dirty {
		return o.GetObjectConvertor().Convert(versionedAttr.VersionedObject, versionedAttr.Attributes.GetObject(), nil)
	}
	return nil
}

// note that callAttrMutatingHook updates attr
func (a *mutatingDispatcher) callAttrMutatingHook(ctx context.Context, invocation *generic.WebhookInvocation, attr *generic.VersionedAttributes, o admission.ObjectInterfaces) error {
	h := invocation.Webhook
	if attr.Attributes.IsDryRun() {
		if h.SideEffects == nil {
			return &webhook.ErrCallingWebhook{WebhookName: h.Name, Reason: fmt.Errorf("Webhook SideEffects is nil")}
		}
		if !(*h.SideEffects == v1beta1.SideEffectClassNone || *h.SideEffects == v1beta1.SideEffectClassNoneOnDryRun) {
			return webhookerrors.NewDryRunUnsupportedErr(h.Name)
		}
	}

	// Currently dispatcher only supports `v1beta1` AdmissionReview
	// TODO: Make the dispatcher capable of sending multiple AdmissionReview versions
	if !util.HasAdmissionReviewVersion(v1beta1.SchemeGroupVersion.Version, h) {
		return &webhook.ErrCallingWebhook{WebhookName: h.Name, Reason: fmt.Errorf("webhook does not accept v1beta1 AdmissionReview")}
	}

	// Make the webhook request
	request := request.CreateAdmissionReview(attr, invocation)
	client, err := a.cm.HookClient(util.HookClientConfigForWebhook(h))
	if err != nil {
		return &webhook.ErrCallingWebhook{WebhookName: h.Name, Reason: err}
	}
	response := &admissionv1beta1.AdmissionReview{}
	r := client.Post().Context(ctx).Body(&request)
	if h.TimeoutSeconds != nil {
		r = r.Timeout(time.Duration(*h.TimeoutSeconds) * time.Second)
	}
	if err := r.Do().Into(response); err != nil {
		return &webhook.ErrCallingWebhook{WebhookName: h.Name, Reason: err}
	}

	if response.Response == nil {
		return &webhook.ErrCallingWebhook{WebhookName: h.Name, Reason: fmt.Errorf("Webhook response was absent")}
	}

	for k, v := range response.Response.AuditAnnotations {
		key := h.Name + "/" + k
		if err := attr.Attributes.AddAnnotation(key, v); err != nil {
			klog.Warningf("Failed to set admission audit annotation %s to %s for mutating webhook %s: %v", key, v, h.Name, err)
		}
	}

	if !response.Response.Allowed {
		return webhookerrors.ToStatusErr(h.Name, response.Response.Result)
	}

	patchJS := response.Response.Patch
	if len(patchJS) == 0 {
		return nil
	}
	patchObj, err := jsonpatch.DecodePatch(patchJS)
	if err != nil {
		return apierrors.NewInternalError(err)
	}
	if len(patchObj) == 0 {
		return nil
	}

	// if a non-empty patch was provided, and we have no object we can apply it to (e.g. a DELETE admission operation), error
	if attr.VersionedObject == nil {
		return apierrors.NewInternalError(fmt.Errorf("admission webhook %q attempted to modify the object, which is not supported for this operation", h.Name))
	}

	jsonSerializer := json.NewSerializer(json.DefaultMetaFactory, o.GetObjectCreater(), o.GetObjectTyper(), false)
	objJS, err := runtime.Encode(jsonSerializer, attr.VersionedObject)
	if err != nil {
		return apierrors.NewInternalError(err)
	}
	patchedJS, err := patchObj.Apply(objJS)
	if err != nil {
		return apierrors.NewInternalError(err)
	}

	var newVersionedObject runtime.Object
	if _, ok := attr.VersionedObject.(*unstructured.Unstructured); ok {
		// Custom Resources don't have corresponding Go struct's.
		// They are represented as Unstructured.
		newVersionedObject = &unstructured.Unstructured{}
	} else {
		newVersionedObject, err = o.GetObjectCreater().New(attr.VersionedKind)
		if err != nil {
			return apierrors.NewInternalError(err)
		}
	}
	// TODO: if we have multiple mutating webhooks, we can remember the json
	// instead of encoding and decoding for each one.
	if newVersionedObject, _, err = jsonSerializer.Decode(patchedJS, nil, newVersionedObject); err != nil {
		return apierrors.NewInternalError(err)
	}
	attr.Dirty = true
	attr.VersionedObject = newVersionedObject
	o.GetObjectDefaulter().Default(attr.VersionedObject)
	return nil
}
