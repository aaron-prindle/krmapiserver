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

package configmap

import (
	"testing"

	metav1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/apis/meta/v1"
	genericapirequest "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/endpoints/request"
	api "github.com/aaron-prindle/krmapiserver/pkg/apis/core"
)

func TestConfigMapStrategy(t *testing.T) {
	ctx := genericapirequest.NewDefaultContext()
	if !Strategy.NamespaceScoped() {
		t.Errorf("ConfigMap must be namespace scoped")
	}
	if Strategy.AllowCreateOnUpdate() {
		t.Errorf("ConfigMap should not allow create on update")
	}

	cfg := &api.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "valid-config-data",
			Namespace: metav1.NamespaceDefault,
		},
		Data: map[string]string{
			"foo": "bar",
		},
	}

	Strategy.PrepareForCreate(ctx, cfg)

	errs := Strategy.Validate(ctx, cfg)
	if len(errs) != 0 {
		t.Errorf("unexpected error validating %v", errs)
	}

	newCfg := &api.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "valid-config-data-2",
			Namespace:       metav1.NamespaceDefault,
			ResourceVersion: "4",
		},
		Data: map[string]string{
			"invalidKey": "updatedValue",
		},
	}

	Strategy.PrepareForUpdate(ctx, newCfg, cfg)

	errs = Strategy.ValidateUpdate(ctx, newCfg, cfg)
	if len(errs) == 0 {
		t.Errorf("Expected a validation error")
	}
}
