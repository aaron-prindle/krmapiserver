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

package validating

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/klog"

	admissionv1beta1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/admission/v1beta1"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/api/admissionregistration/v1beta1"
	apierrors "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/api/errors"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/admission"
	admissionmetrics "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/admission/metrics"
	webhookerrors "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/admission/plugin/webhook/errors"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/admission/plugin/webhook/generic"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/admission/plugin/webhook/request"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/admission/plugin/webhook/util"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/util/webhook"
)

type validatingDispatcher struct {
	cm *webhook.ClientManager
}

func newValidatingDispatcher(cm *webhook.ClientManager) generic.Dispatcher {
	return &validatingDispatcher{cm}
}

var _ generic.Dispatcher = &validatingDispatcher{}

func (d *validatingDispatcher) Dispatch(ctx context.Context, attr admission.Attributes, o admission.ObjectInterfaces, relevantHooks []*generic.WebhookInvocation) error {
	// Construct all the versions we need to call our webhooks
	versionedAttrs := map[schema.GroupVersionKind]*generic.VersionedAttributes{}
	for _, call := range relevantHooks {
		// If we already have this version, continue
		if _, ok := versionedAttrs[call.Kind]; ok {
			continue
		}
		versionedAttr, err := generic.NewVersionedAttributes(attr, call.Kind, o)
		if err != nil {
			return apierrors.NewInternalError(err)
		}
		versionedAttrs[call.Kind] = versionedAttr
	}

	wg := sync.WaitGroup{}
	errCh := make(chan error, len(relevantHooks))
	wg.Add(len(relevantHooks))
	for i := range relevantHooks {
		go func(invocation *generic.WebhookInvocation) {
			defer wg.Done()
			hook := invocation.Webhook
			versionedAttr := versionedAttrs[invocation.Kind]

			t := time.Now()
			err := d.callHook(ctx, invocation, versionedAttr)
			admissionmetrics.Metrics.ObserveWebhook(time.Since(t), err != nil, versionedAttr.Attributes, "validating", hook.Name)
			if err == nil {
				return
			}

			ignoreClientCallFailures := hook.FailurePolicy != nil && *hook.FailurePolicy == v1beta1.Ignore
			if callErr, ok := err.(*webhook.ErrCallingWebhook); ok {
				if ignoreClientCallFailures {
					klog.Warningf("Failed calling webhook, failing open %v: %v", hook.Name, callErr)
					utilruntime.HandleError(callErr)
					return
				}

				klog.Warningf("Failed calling webhook, failing closed %v: %v", hook.Name, err)
				errCh <- apierrors.NewInternalError(err)
				return
			}

			klog.Warningf("rejected by webhook %q: %#v", hook.Name, err)
			errCh <- err
		}(relevantHooks[i])
	}
	wg.Wait()
	close(errCh)

	var errs []error
	for e := range errCh {
		errs = append(errs, e)
	}
	if len(errs) == 0 {
		return nil
	}
	if len(errs) > 1 {
		for i := 1; i < len(errs); i++ {
			// TODO: merge status errors; until then, just return the first one.
			utilruntime.HandleError(errs[i])
		}
	}
	return errs[0]
}

func (d *validatingDispatcher) callHook(ctx context.Context, invocation *generic.WebhookInvocation, attr *generic.VersionedAttributes) error {
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
		return &webhook.ErrCallingWebhook{WebhookName: h.Name, Reason: fmt.Errorf("webhook does not accept v1beta1 AdmissionReviewRequest")}
	}

	// Make the webhook request
	request := request.CreateAdmissionReview(attr, invocation)
	client, err := d.cm.HookClient(util.HookClientConfigForWebhook(h))
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
			klog.Warningf("Failed to set admission audit annotation %s to %s for validating webhook %s: %v", key, v, h.Name, err)
		}
	}
	if response.Response.Allowed {
		return nil
	}
	return webhookerrors.ToStatusErr(h.Name, response.Response.Result)
}
