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

package v1_test

import (
	"reflect"
	"testing"

	autoscalingv1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/autoscaling/v1"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	"github.com/aaron-prindle/krmapiserver/pkg/api/legacyscheme"
	_ "github.com/aaron-prindle/krmapiserver/pkg/apis/autoscaling/install"
	. "github.com/aaron-prindle/krmapiserver/pkg/apis/autoscaling/v1"
	_ "github.com/aaron-prindle/krmapiserver/pkg/apis/core/install"
	utilpointer "github.com/aaron-prindle/krmapiserver/included/k8s.io/utils/pointer"
)

func TestSetDefaultHPA(t *testing.T) {
	tests := []struct {
		hpa            autoscalingv1.HorizontalPodAutoscaler
		expectReplicas int32
		test           string
	}{
		{
			hpa:            autoscalingv1.HorizontalPodAutoscaler{},
			expectReplicas: 1,
			test:           "unspecified min replicas, use the default value",
		},
		{
			hpa: autoscalingv1.HorizontalPodAutoscaler{
				Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
					MinReplicas: utilpointer.Int32Ptr(3),
				},
			},
			expectReplicas: 3,
			test:           "set min replicas to 3",
		},
	}

	for _, test := range tests {
		hpa := &test.hpa
		obj2 := roundTrip(t, runtime.Object(hpa))
		hpa2, ok := obj2.(*autoscalingv1.HorizontalPodAutoscaler)
		if !ok {
			t.Fatalf("unexpected object: %v", obj2)
		}
		if hpa2.Spec.MinReplicas == nil {
			t.Errorf("unexpected nil MinReplicas")
		} else if test.expectReplicas != *hpa2.Spec.MinReplicas {
			t.Errorf("expected: %d MinReplicas, got: %d", test.expectReplicas, *hpa2.Spec.MinReplicas)
		}
	}
}

func roundTrip(t *testing.T, obj runtime.Object) runtime.Object {
	data, err := runtime.Encode(legacyscheme.Codecs.LegacyCodec(SchemeGroupVersion), obj)
	if err != nil {
		t.Errorf("%v\n %#v", err, obj)
		return nil
	}
	obj2, err := runtime.Decode(legacyscheme.Codecs.UniversalDecoder(), data)
	if err != nil {
		t.Errorf("%v\nData: %s\nSource: %#v", err, string(data), obj)
		return nil
	}
	obj3 := reflect.New(reflect.TypeOf(obj).Elem()).Interface().(runtime.Object)
	err = legacyscheme.Scheme.Convert(obj2, obj3, nil)
	if err != nil {
		t.Errorf("%v\nSource: %#v", err, obj2)
		return nil
	}
	return obj3
}
