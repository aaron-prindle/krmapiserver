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

package openapi

import (
	"reflect"
	"strings"
	"testing"

	"github.com/aaron-prindle/krmapiserver/included/github.com/go-openapi/spec"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	openapitesting "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/endpoints/openapi/testing"
)

func assertEqual(t *testing.T, expected, actual interface{}) {
	var equal bool
	if expected == nil || actual == nil {
		equal = expected == actual
	} else {
		equal = reflect.DeepEqual(expected, actual)
	}
	if !equal {
		t.Errorf("%v != %v", expected, actual)
	}
}

func TestGetDefinitionName(t *testing.T) {
	testType := openapitesting.TestType{}
	// in production, the name is stripped of ".*included/" prefix before passed
	// to GetDefinitionName, so here typePkgName does not have the
	// "github.com/aaron-prindle/krmapiserver/included" prefix.
	typePkgName := "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/endpoints/openapi/testing.TestType"
	typeFriendlyName := "io.k8s.apiserver.pkg.endpoints.openapi.testing.TestType"
	if strings.HasSuffix(reflect.TypeOf(testType).PkgPath(), "go_default_test") {
		// the test is running inside bazel where the package name is changed and
		// "go_default_test" will add to package path.
		typePkgName = "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/endpoints/openapi/testing/go_default_test.TestType"
		typeFriendlyName = "io.k8s.apiserver.pkg.endpoints.openapi.testing.go_default_test.TestType"
	}
	s := runtime.NewScheme()
	s.AddKnownTypeWithName(testType.GroupVersionKind(), &testType)
	namer := NewDefinitionNamer(s)
	n, e := namer.GetDefinitionName(typePkgName)
	assertEqual(t, typeFriendlyName, n)
	assertEqual(t, []interface{}{
		map[string]interface{}{
			"group":   "test",
			"version": "v1",
			"kind":    "TestType",
		},
	}, e["x-kubernetes-group-version-kind"])
	n, e2 := namer.GetDefinitionName("test.com/another.Type")
	assertEqual(t, "com.test.another.Type", n)
	assertEqual(t, e2, spec.Extensions(nil))
}
