/*
Copyright 2016 The Kubernetes Authors.

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

// Package webhook implements the authorizer.Authorizer interface using HTTP webhooks.
package webhook

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/klog"

	authorization "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/authorization/v1beta1"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime/schema"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/cache"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/authentication/user"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/authorization/authorizer"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/util/webhook"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/kubernetes/scheme"
	authorizationclient "github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/kubernetes/typed/authorization/v1beta1"
)

var (
	groupVersions = []schema.GroupVersion{authorization.SchemeGroupVersion}
)

const (
	retryBackoff = 500 * time.Millisecond
	// The maximum length of requester-controlled attributes to allow caching.
	maxControlledAttrCacheSize = 10000
)

// Ensure Webhook implements the authorizer.Authorizer interface.
var _ authorizer.Authorizer = (*WebhookAuthorizer)(nil)

type WebhookAuthorizer struct {
	subjectAccessReview authorizationclient.SubjectAccessReviewInterface
	responseCache       *cache.LRUExpireCache
	authorizedTTL       time.Duration
	unauthorizedTTL     time.Duration
	initialBackoff      time.Duration
	decisionOnError     authorizer.Decision
}

// NewFromInterface creates a WebhookAuthorizer using the given subjectAccessReview client
func NewFromInterface(subjectAccessReview authorizationclient.SubjectAccessReviewInterface, authorizedTTL, unauthorizedTTL time.Duration) (*WebhookAuthorizer, error) {
	return newWithBackoff(subjectAccessReview, authorizedTTL, unauthorizedTTL, retryBackoff)
}

// New creates a new WebhookAuthorizer from the provided kubeconfig file.
//
// The config's cluster field is used to refer to the remote service, user refers to the returned authorizer.
//
//     # clusters refers to the remote service.
//     clusters:
//     - name: name-of-remote-authz-service
//       cluster:
//         certificate-authority: /path/to/ca.pem      # CA for verifying the remote service.
//         server: https://authz.example.com/authorize # URL of remote service to query. Must use 'https'.
//
//     # users refers to the API server's webhook configuration.
//     users:
//     - name: name-of-api-server
//       user:
//         client-certificate: /path/to/cert.pem # cert for the webhook plugin to use
//         client-key: /path/to/key.pem          # key matching the cert
//
// For additional HTTP configuration, refer to the kubeconfig documentation
// https://kubernetes.io/docs/user-guide/kubeconfig-file/.
func New(kubeConfigFile string, authorizedTTL, unauthorizedTTL time.Duration) (*WebhookAuthorizer, error) {
	subjectAccessReview, err := subjectAccessReviewInterfaceFromKubeconfig(kubeConfigFile)
	if err != nil {
		return nil, err
	}
	return newWithBackoff(subjectAccessReview, authorizedTTL, unauthorizedTTL, retryBackoff)
}

// newWithBackoff allows tests to skip the sleep.
func newWithBackoff(subjectAccessReview authorizationclient.SubjectAccessReviewInterface, authorizedTTL, unauthorizedTTL, initialBackoff time.Duration) (*WebhookAuthorizer, error) {
	return &WebhookAuthorizer{
		subjectAccessReview: subjectAccessReview,
		responseCache:       cache.NewLRUExpireCache(1024),
		authorizedTTL:       authorizedTTL,
		unauthorizedTTL:     unauthorizedTTL,
		initialBackoff:      initialBackoff,
		decisionOnError:     authorizer.DecisionNoOpinion,
	}, nil
}

// Authorize makes a REST request to the remote service describing the attempted action as a JSON
// serialized api.authorization.v1beta1.SubjectAccessReview object. An example request body is
// provided below.
//
//     {
//       "apiVersion": "authorization.k8s.io/v1beta1",
//       "kind": "SubjectAccessReview",
//       "spec": {
//         "resourceAttributes": {
//           "namespace": "kittensandponies",
//           "verb": "GET",
//           "group": "group3",
//           "resource": "pods"
//         },
//         "user": "jane",
//         "group": [
//           "group1",
//           "group2"
//         ]
//       }
//     }
//
// The remote service is expected to fill the SubjectAccessReviewStatus field to either allow or
// disallow access. A permissive response would return:
//
//     {
//       "apiVersion": "authorization.k8s.io/v1beta1",
//       "kind": "SubjectAccessReview",
//       "status": {
//         "allowed": true
//       }
//     }
//
// To disallow access, the remote service would return:
//
//     {
//       "apiVersion": "authorization.k8s.io/v1beta1",
//       "kind": "SubjectAccessReview",
//       "status": {
//         "allowed": false,
//         "reason": "user does not have read access to the namespace"
//       }
//     }
//
// TODO(mikedanese): We should eventually support failing closed when we
// encounter an error. We are failing open now to preserve backwards compatible
// behavior.
func (w *WebhookAuthorizer) Authorize(attr authorizer.Attributes) (decision authorizer.Decision, reason string, err error) {
	r := &authorization.SubjectAccessReview{}
	if user := attr.GetUser(); user != nil {
		r.Spec = authorization.SubjectAccessReviewSpec{
			User:   user.GetName(),
			UID:    user.GetUID(),
			Groups: user.GetGroups(),
			Extra:  convertToSARExtra(user.GetExtra()),
		}
	}

	if attr.IsResourceRequest() {
		r.Spec.ResourceAttributes = &authorization.ResourceAttributes{
			Namespace:   attr.GetNamespace(),
			Verb:        attr.GetVerb(),
			Group:       attr.GetAPIGroup(),
			Version:     attr.GetAPIVersion(),
			Resource:    attr.GetResource(),
			Subresource: attr.GetSubresource(),
			Name:        attr.GetName(),
		}
	} else {
		r.Spec.NonResourceAttributes = &authorization.NonResourceAttributes{
			Path: attr.GetPath(),
			Verb: attr.GetVerb(),
		}
	}
	key, err := json.Marshal(r.Spec)
	if err != nil {
		return w.decisionOnError, "", err
	}
	if entry, ok := w.responseCache.Get(string(key)); ok {
		r.Status = entry.(authorization.SubjectAccessReviewStatus)
	} else {
		var (
			result *authorization.SubjectAccessReview
			err    error
		)
		webhook.WithExponentialBackoff(w.initialBackoff, func() error {
			result, err = w.subjectAccessReview.Create(r)
			return err
		})
		if err != nil {
			// An error here indicates bad configuration or an outage. Log for debugging.
			klog.Errorf("Failed to make webhook authorizer request: %v", err)
			return w.decisionOnError, "", err
		}
		r.Status = result.Status
		if shouldCache(attr) {
			if r.Status.Allowed {
				w.responseCache.Add(string(key), r.Status, w.authorizedTTL)
			} else {
				w.responseCache.Add(string(key), r.Status, w.unauthorizedTTL)
			}
		}
	}
	switch {
	case r.Status.Denied && r.Status.Allowed:
		return authorizer.DecisionDeny, r.Status.Reason, fmt.Errorf("webhook subject access review returned both allow and deny response")
	case r.Status.Denied:
		return authorizer.DecisionDeny, r.Status.Reason, nil
	case r.Status.Allowed:
		return authorizer.DecisionAllow, r.Status.Reason, nil
	default:
		return authorizer.DecisionNoOpinion, r.Status.Reason, nil
	}

}

//TODO: need to finish the method to get the rules when using webhook mode
func (w *WebhookAuthorizer) RulesFor(user user.Info, namespace string) ([]authorizer.ResourceRuleInfo, []authorizer.NonResourceRuleInfo, bool, error) {
	var (
		resourceRules    []authorizer.ResourceRuleInfo
		nonResourceRules []authorizer.NonResourceRuleInfo
	)
	incomplete := true
	return resourceRules, nonResourceRules, incomplete, fmt.Errorf("webhook authorizer does not support user rule resolution")
}

func convertToSARExtra(extra map[string][]string) map[string]authorization.ExtraValue {
	if extra == nil {
		return nil
	}
	ret := map[string]authorization.ExtraValue{}
	for k, v := range extra {
		ret[k] = authorization.ExtraValue(v)
	}

	return ret
}

// subjectAccessReviewInterfaceFromKubeconfig builds a client from the specified kubeconfig file,
// and returns a SubjectAccessReviewInterface that uses that client. Note that the client submits SubjectAccessReview
// requests to the exact path specified in the kubeconfig file, so arbitrary non-API servers can be targeted.
func subjectAccessReviewInterfaceFromKubeconfig(kubeConfigFile string) (authorizationclient.SubjectAccessReviewInterface, error) {
	localScheme := runtime.NewScheme()
	if err := scheme.AddToScheme(localScheme); err != nil {
		return nil, err
	}
	if err := localScheme.SetVersionPriority(groupVersions...); err != nil {
		return nil, err
	}

	gw, err := webhook.NewGenericWebhook(localScheme, scheme.Codecs, kubeConfigFile, groupVersions, 0)
	if err != nil {
		return nil, err
	}
	return &subjectAccessReviewClient{gw}, nil
}

type subjectAccessReviewClient struct {
	w *webhook.GenericWebhook
}

func (t *subjectAccessReviewClient) Create(subjectAccessReview *authorization.SubjectAccessReview) (*authorization.SubjectAccessReview, error) {
	result := &authorization.SubjectAccessReview{}
	err := t.w.RestClient.Post().Body(subjectAccessReview).Do().Into(result)
	return result, err
}

// shouldCache determines whether it is safe to cache the given request attributes. If the
// requester-controlled attributes are too large, this may be a DoS attempt, so we skip the cache.
func shouldCache(attr authorizer.Attributes) bool {
	controlledAttrSize := int64(len(attr.GetNamespace())) +
		int64(len(attr.GetVerb())) +
		int64(len(attr.GetAPIGroup())) +
		int64(len(attr.GetAPIVersion())) +
		int64(len(attr.GetResource())) +
		int64(len(attr.GetSubresource())) +
		int64(len(attr.GetName())) +
		int64(len(attr.GetPath()))
	return controlledAttrSize < maxControlledAttrCacheSize
}
