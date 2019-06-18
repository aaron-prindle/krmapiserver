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

// +k8s:conversion-gen=k8s.io/kubernetes/pkg/apis/storage
// +k8s:conversion-gen-external-types=k8s.io/api/storage/v1alpha1
// +groupName=storage.k8s.io
// +k8s:defaulter-gen=TypeMeta
// +k8s:defaulter-gen-input=../../../../included/k8s.io/api/storage/v1alpha1

package v1alpha1 // import "github.com/aaron-prindle/krmapiserver/pkg/apis/storage/v1alpha1"
