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

package openapi

import (
	"github.com/aaron-prindle/krmapiserver/included/github.com/go-openapi/spec"
)

// mergeSpecs aggregates all OpenAPI specs, reusing the metadata of the first, static spec as the basis.
func mergeSpecs(staticSpec *spec.Swagger, crdSpecs ...*spec.Swagger) *spec.Swagger {
	// create shallow copy of staticSpec, but replace paths and definitions because we modify them.
	specToReturn := *staticSpec
	if staticSpec.Definitions != nil {
		specToReturn.Definitions = spec.Definitions{}
		for k, s := range staticSpec.Definitions {
			specToReturn.Definitions[k] = s
		}
	}
	if staticSpec.Paths != nil {
		specToReturn.Paths = &spec.Paths{
			Paths: map[string]spec.PathItem{},
		}
		for k, p := range staticSpec.Paths.Paths {
			specToReturn.Paths.Paths[k] = p
		}
	}

	for _, s := range crdSpecs {
		mergeSpec(&specToReturn, s)
	}

	return &specToReturn
}

// mergeSpec copies paths and definitions from source to dest, mutating dest, but not source.
// We assume that conflicts do not matter.
func mergeSpec(dest, source *spec.Swagger) {
	if source == nil || source.Paths == nil {
		return
	}
	if dest.Paths == nil {
		dest.Paths = &spec.Paths{}
	}
	for k, v := range source.Definitions {
		if dest.Definitions == nil {
			dest.Definitions = spec.Definitions{}
		}
		dest.Definitions[k] = v
	}
	for k, v := range source.Paths.Paths {
		if dest.Paths.Paths == nil {
			dest.Paths.Paths = map[string]spec.PathItem{}
		}
		dest.Paths.Paths[k] = v
	}
}
