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

package node

import (
	"testing"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/fields"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/labels"
	apitesting "github.com/aaron-prindle/krmapiserver/pkg/api/testing"
	api "github.com/aaron-prindle/krmapiserver/pkg/apis/core"

	// install all api groups for testing
	_ "github.com/aaron-prindle/krmapiserver/pkg/api/testapi"
)

func TestMatchNode(t *testing.T) {
	testFieldMap := map[bool][]fields.Set{
		true: {
			{"metadata.name": "foo"},
		},
		false: {
			{"foo": "bar"},
		},
	}

	for expectedResult, fieldSet := range testFieldMap {
		for _, field := range fieldSet {
			m := MatchNode(labels.Everything(), field.AsSelector())
			_, matchesSingle := m.MatchesSingle()
			if e, a := expectedResult, matchesSingle; e != a {
				t.Errorf("%+v: expected %v, got %v", fieldSet, e, a)
			}
		}
	}
}

func TestSelectableFieldLabelConversions(t *testing.T) {
	apitesting.TestSelectableFieldLabelConversionsOfKind(t,
		"v1",
		"Node",
		NodeToSelectableFields(&api.Node{}),
		nil,
	)
}
