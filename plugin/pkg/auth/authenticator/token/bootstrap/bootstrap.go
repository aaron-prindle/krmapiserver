/*
Copyright 2017 The Kubernetes Authors.

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

/*
Package bootstrap provides a token authenticator for TLS bootstrap secrets.
*/
package bootstrap

import (
	"context"
	"crypto/subtle"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/klog"

	corev1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/core/v1"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/api/errors"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/sets"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/authentication/authenticator"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/authentication/user"
	corev1listers "github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/listers/core/v1"
	bootstrapapi "github.com/aaron-prindle/krmapiserver/included/k8s.io/cluster-bootstrap/token/api"
	bootstraputil "github.com/aaron-prindle/krmapiserver/included/k8s.io/cluster-bootstrap/token/util"
)

// TODO: A few methods in this package is copied from other sources. Either
// because the existing functionality isn't exported or because it is in a
// package that shouldn't be directly imported by this packages.

// NewTokenAuthenticator initializes a bootstrap token authenticator.
//
// Lister is expected to be for the "kube-system" namespace.
func NewTokenAuthenticator(lister corev1listers.SecretNamespaceLister) *TokenAuthenticator {
	return &TokenAuthenticator{lister}
}

// TokenAuthenticator authenticates bootstrap tokens from secrets in the API server.
type TokenAuthenticator struct {
	lister corev1listers.SecretNamespaceLister
}

// tokenErrorf prints a error message for a secret that has matched a bearer
// token but fails to meet some other criteria.
//
//    tokenErrorf(secret, "has invalid value for key %s", key)
//
func tokenErrorf(s *corev1.Secret, format string, i ...interface{}) {
	format = fmt.Sprintf("Bootstrap secret %s/%s matching bearer token ", s.Namespace, s.Name) + format
	klog.V(3).Infof(format, i...)
}

// AuthenticateToken tries to match the provided token to a bootstrap token secret
// in a given namespace. If found, it authenticates the token in the
// "system:bootstrappers" group and with the "system:bootstrap:(token-id)" username.
//
// All secrets must be of type "bootstrap.kubernetes.io/token". An example secret:
//
//     apiVersion: v1
//     kind: Secret
//     metadata:
//       # Name MUST be of form "bootstrap-token-( token id )".
//       name: bootstrap-token-( token id )
//       namespace: kube-system
//     # Only secrets of this type will be evaluated.
//     type: bootstrap.kubernetes.io/token
//     data:
//       token-secret: ( private part of token )
//       token-id: ( token id )
//       # Required key usage.
//       usage-bootstrap-authentication: true
//       auth-extra-groups: "system:bootstrappers:custom-group1,system:bootstrappers:custom-group2"
//       # May also contain an expiry.
//
// Tokens are expected to be of the form:
//
//     ( token-id ).( token-secret )
//
func (t *TokenAuthenticator) AuthenticateToken(ctx context.Context, token string) (*authenticator.Response, bool, error) {
	tokenID, tokenSecret, err := parseToken(token)
	if err != nil {
		// Token isn't of the correct form, ignore it.
		return nil, false, nil
	}

	secretName := bootstrapapi.BootstrapTokenSecretPrefix + tokenID
	secret, err := t.lister.Get(secretName)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.V(3).Infof("No secret of name %s to match bootstrap bearer token", secretName)
			return nil, false, nil
		}
		return nil, false, err
	}

	if secret.DeletionTimestamp != nil {
		tokenErrorf(secret, "is deleted and awaiting removal")
		return nil, false, nil
	}

	if string(secret.Type) != string(bootstrapapi.SecretTypeBootstrapToken) || secret.Data == nil {
		tokenErrorf(secret, "has invalid type, expected %s.", bootstrapapi.SecretTypeBootstrapToken)
		return nil, false, nil
	}

	ts := getSecretString(secret, bootstrapapi.BootstrapTokenSecretKey)
	if subtle.ConstantTimeCompare([]byte(ts), []byte(tokenSecret)) != 1 {
		tokenErrorf(secret, "has invalid value for key %s, expected %s.", bootstrapapi.BootstrapTokenSecretKey, tokenSecret)
		return nil, false, nil
	}

	id := getSecretString(secret, bootstrapapi.BootstrapTokenIDKey)
	if id != tokenID {
		tokenErrorf(secret, "has invalid value for key %s, expected %s.", bootstrapapi.BootstrapTokenIDKey, tokenID)
		return nil, false, nil
	}

	if isSecretExpired(secret) {
		// logging done in isSecretExpired method.
		return nil, false, nil
	}

	if getSecretString(secret, bootstrapapi.BootstrapTokenUsageAuthentication) != "true" {
		tokenErrorf(secret, "not marked %s=true.", bootstrapapi.BootstrapTokenUsageAuthentication)
		return nil, false, nil
	}

	groups, err := getGroups(secret)
	if err != nil {
		tokenErrorf(secret, "has invalid value for key %s: %v.", bootstrapapi.BootstrapTokenExtraGroupsKey, err)
		return nil, false, nil
	}

	return &authenticator.Response{
		User: &user.DefaultInfo{
			Name:   bootstrapapi.BootstrapUserPrefix + string(id),
			Groups: groups,
		},
	}, true, nil
}

// Copied from k8s.io/cluster-bootstrap/token/api
func getSecretString(secret *corev1.Secret, key string) string {
	data, ok := secret.Data[key]
	if !ok {
		return ""
	}

	return string(data)
}

// Copied from k8s.io/cluster-bootstrap/token/api
func isSecretExpired(secret *corev1.Secret) bool {
	expiration := getSecretString(secret, bootstrapapi.BootstrapTokenExpirationKey)
	if len(expiration) > 0 {
		expTime, err2 := time.Parse(time.RFC3339, expiration)
		if err2 != nil {
			klog.V(3).Infof("Unparseable expiration time (%s) in %s/%s Secret: %v. Treating as expired.",
				expiration, secret.Namespace, secret.Name, err2)
			return true
		}
		if time.Now().After(expTime) {
			klog.V(3).Infof("Expired bootstrap token in %s/%s Secret: %v",
				secret.Namespace, secret.Name, expiration)
			return true
		}
	}
	return false
}

// Copied from kubernetes/cmd/kubeadm/app/util/token

var (
	// tokenRegexpString defines id.secret regular expression pattern
	tokenRegexpString = "^([a-z0-9]{6})\\.([a-z0-9]{16})$"
	// tokenRegexp is a compiled regular expression of TokenRegexpString
	tokenRegexp = regexp.MustCompile(tokenRegexpString)
)

// parseToken tries and parse a valid token from a string.
// A token ID and token secret are returned in case of success, an error otherwise.
func parseToken(s string) (string, string, error) {
	split := tokenRegexp.FindStringSubmatch(s)
	if len(split) != 3 {
		return "", "", fmt.Errorf("token [%q] was not of form [%q]", s, tokenRegexpString)
	}
	return split[1], split[2], nil
}

// getGroups loads and validates the bootstrapapi.BootstrapTokenExtraGroupsKey
// key from the bootstrap token secret, returning a list of group names or an
// error if any of the group names are invalid.
func getGroups(secret *corev1.Secret) ([]string, error) {
	// always include the default group
	groups := sets.NewString(bootstrapapi.BootstrapDefaultGroup)

	// grab any extra groups and if there are none, return just the default
	extraGroupsString := getSecretString(secret, bootstrapapi.BootstrapTokenExtraGroupsKey)
	if extraGroupsString == "" {
		return groups.List(), nil
	}

	// validate the names of the extra groups
	for _, group := range strings.Split(extraGroupsString, ",") {
		if err := bootstraputil.ValidateBootstrapGroupName(group); err != nil {
			return nil, err
		}
		groups.Insert(group)
	}

	// return the result as a deduplicated, sorted list
	return groups.List(), nil
}
