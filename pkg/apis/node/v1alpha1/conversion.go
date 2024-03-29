/*
Copyright 2019 The Kubernetes Authors.

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

package v1alpha1

import (
	v1alpha1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/node/v1alpha1"
	conversion "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/conversion"
	runtime "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	node "github.com/aaron-prindle/krmapiserver/pkg/apis/node"
)

func addConversionFuncs(s *runtime.Scheme) error {
	return s.AddConversionFuncs(
		Convert_v1alpha1_RuntimeClass_To_node_RuntimeClass,
		Convert_node_RuntimeClass_To_v1alpha1_RuntimeClass,
	)
}

func Convert_v1alpha1_RuntimeClass_To_node_RuntimeClass(in *v1alpha1.RuntimeClass, out *node.RuntimeClass, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	out.Handler = in.Spec.RuntimeHandler
	if in.Spec.Scheduling != nil {
		out.Scheduling = new(node.Scheduling)
		autoConvert_v1alpha1_Scheduling_To_node_Scheduling(in.Spec.Scheduling, out.Scheduling, s)
	}
	return nil
}

func Convert_node_RuntimeClass_To_v1alpha1_RuntimeClass(in *node.RuntimeClass, out *v1alpha1.RuntimeClass, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	out.Spec.RuntimeHandler = in.Handler
	if in.Scheduling != nil {
		out.Spec.Scheduling = new(v1alpha1.Scheduling)
		autoConvert_node_Scheduling_To_v1alpha1_Scheduling(in.Scheduling, out.Spec.Scheduling, s)
	}
	return nil
}
