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

package authenticatorfactory

import (
	"errors"
	"fmt"
	"time"

	"github.com/aaron-prindle/krmapiserver/included/github.com/go-openapi/spec"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/authentication/authenticator"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/authentication/group"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/authentication/request/anonymous"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/authentication/request/bearertoken"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/authentication/request/headerrequest"
	unionauth "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/authentication/request/union"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/authentication/request/websocket"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/authentication/request/x509"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/authentication/token/cache"
	webhooktoken "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/plugin/pkg/authenticator/token/webhook"
	authenticationclient "github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/kubernetes/typed/authentication/v1beta1"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/util/cert"
)

// DelegatingAuthenticatorConfig is the minimal configuration needed to create an authenticator
// built to delegate authentication to a kube API server
type DelegatingAuthenticatorConfig struct {
	Anonymous bool

	// TokenAccessReviewClient is a client to do token review. It can be nil. Then every token is ignored.
	TokenAccessReviewClient authenticationclient.TokenReviewInterface

	// CacheTTL is the length of time that a token authentication answer will be cached.
	CacheTTL time.Duration

	// ClientCAFile is the CA bundle file used to authenticate client certificates
	ClientCAFile string

	APIAudiences authenticator.Audiences

	RequestHeaderConfig *RequestHeaderConfig
}

func (c DelegatingAuthenticatorConfig) New() (authenticator.Request, *spec.SecurityDefinitions, error) {
	authenticators := []authenticator.Request{}
	securityDefinitions := spec.SecurityDefinitions{}

	// front-proxy first, then remote
	// Add the front proxy authenticator if requested
	if c.RequestHeaderConfig != nil {
		requestHeaderAuthenticator, err := headerrequest.NewSecure(
			c.RequestHeaderConfig.ClientCA,
			c.RequestHeaderConfig.AllowedClientNames,
			c.RequestHeaderConfig.UsernameHeaders,
			c.RequestHeaderConfig.GroupHeaders,
			c.RequestHeaderConfig.ExtraHeaderPrefixes,
		)
		if err != nil {
			return nil, nil, err
		}
		authenticators = append(authenticators, requestHeaderAuthenticator)
	}

	// x509 client cert auth
	if len(c.ClientCAFile) > 0 {
		clientCAs, err := cert.NewPool(c.ClientCAFile)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to load client CA file %s: %v", c.ClientCAFile, err)
		}
		verifyOpts := x509.DefaultVerifyOptions()
		verifyOpts.Roots = clientCAs
		authenticators = append(authenticators, x509.New(verifyOpts, x509.CommonNameUserConversion))
	}

	if c.TokenAccessReviewClient != nil {
		tokenAuth, err := webhooktoken.NewFromInterface(c.TokenAccessReviewClient, c.APIAudiences)
		if err != nil {
			return nil, nil, err
		}
		cachingTokenAuth := cache.New(tokenAuth, false, c.CacheTTL, c.CacheTTL)
		authenticators = append(authenticators, bearertoken.New(cachingTokenAuth), websocket.NewProtocolAuthenticator(cachingTokenAuth))

		securityDefinitions["BearerToken"] = &spec.SecurityScheme{
			SecuritySchemeProps: spec.SecuritySchemeProps{
				Type:        "apiKey",
				Name:        "authorization",
				In:          "header",
				Description: "Bearer Token authentication",
			},
		}
	}

	if len(authenticators) == 0 {
		if c.Anonymous {
			return anonymous.NewAuthenticator(), &securityDefinitions, nil
		}
		return nil, nil, errors.New("No authentication method configured")
	}

	authenticator := group.NewAuthenticatedGroupAdder(unionauth.New(authenticators...))
	if c.Anonymous {
		authenticator = unionauth.NewFailOnError(authenticator, anonymous.NewAuthenticator())
	}
	return authenticator, &securityDefinitions, nil
}
