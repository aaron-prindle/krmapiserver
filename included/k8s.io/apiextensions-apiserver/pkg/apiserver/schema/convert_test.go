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

package schema

import (
	"math/rand"
	"reflect"
	"testing"
	"time"

	fuzz "github.com/aaron-prindle/krmapiserver/included/github.com/google/gofuzz"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1beta1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/diff"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/json"
)

func TestStructuralRoundtripOrError(t *testing.T) {
	f := fuzz.New()
	seed := time.Now().UnixNano()
	t.Logf("seed = %v", seed)
	//seed = int64(1549012506261785182)
	f.RandSource(rand.New(rand.NewSource(seed)))
	f.Funcs(
		func(s *apiextensions.JSON, c fuzz.Continue) {
			*s = apiextensions.JSON(map[string]interface{}{"foo": float64(42.2)})
		},
		func(s *apiextensions.JSONSchemaPropsOrArray, c fuzz.Continue) {
			c.FuzzNoCustom(s)
			if s.Schema != nil {
				s.JSONSchemas = nil
			} else if s.JSONSchemas == nil {
				s.Schema = &apiextensions.JSONSchemaProps{}
			}
		},
		func(s *apiextensions.JSONSchemaPropsOrBool, c fuzz.Continue) {
			c.FuzzNoCustom(s)
			if s.Schema != nil {
				s.Allows = false
			}
		},
		func(s **string, c fuzz.Continue) {
			c.FuzzNoCustom(s)
			if *s != nil && **s == "" {
				*s = nil
			}
		},
	)

	f.MaxDepth(2)
	f.NilChance(0.5)

	for i := 0; i < 10000; i++ {
		// fuzz a random field in JSONSchemaProps
		origSchema := &apiextensions.JSONSchemaProps{}
		x := reflect.ValueOf(origSchema).Elem()
		n := rand.Intn(x.NumField())
		if name := x.Type().Field(n).Name; name == "Example" || name == "ExternalDocs" {
			// we drop these intentionally
			continue
		}
		f.Fuzz(x.Field(n).Addr().Interface())
		if origSchema.Nullable {
			// non-empty type for nullable. nullable:true with empty type does not roundtrip because
			// go-openapi does not allow to encode that (we use type slices otherwise).
			origSchema.Type = "string"
		}

		// it roundtrips or NewStructural errors out. We should never drop anything
		orig, err := NewStructural(origSchema)
		if err != nil {
			continue
		}

		// roundtrip through go-openapi, JSON, v1beta1 JSONSchemaProp, internal JSONSchemaProp
		goOpenAPI := orig.ToGoOpenAPI()
		bs, err := json.Marshal(goOpenAPI)
		if err != nil {
			t.Fatal(err)
		}
		str := nullTypeRE.ReplaceAllString(string(bs), `"type":"$1","nullable":true`) // unfold nullable type:[<type>,"null"] -> type:<type>,nullable:true
		v1beta1Schema := &apiextensionsv1beta1.JSONSchemaProps{}
		err = json.Unmarshal([]byte(str), v1beta1Schema)
		if err != nil {
			t.Fatal(err)
		}
		internalSchema := &apiextensions.JSONSchemaProps{}
		err = apiextensionsv1beta1.Convert_v1beta1_JSONSchemaProps_To_apiextensions_JSONSchemaProps(v1beta1Schema, internalSchema, nil)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(origSchema, internalSchema) {
			t.Fatalf("original and result differ: %v", diff.ObjectDiff(origSchema, internalSchema))
		}
	}
}
