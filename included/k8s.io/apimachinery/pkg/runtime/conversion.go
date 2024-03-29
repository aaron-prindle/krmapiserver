/*
Copyright 2014 The Kubernetes Authors.

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

// Package runtime defines conversions between generic types and structs to map query strings
// to struct objects.
package runtime

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/conversion"
)

// DefaultMetaV1FieldSelectorConversion auto-accepts metav1 values for name and namespace.
// A cluster scoped resource specifying namespace empty works fine and specifying a particular
// namespace will return no results, as expected.
func DefaultMetaV1FieldSelectorConversion(label, value string) (string, string, error) {
	switch label {
	case "metadata.name":
		return label, value, nil
	case "metadata.namespace":
		return label, value, nil
	default:
		return "", "", fmt.Errorf("%q is not a known field selector: only %q, %q", label, "metadata.name", "metadata.namespace")
	}
}

// JSONKeyMapper uses the struct tags on a conversion to determine the key value for
// the other side. Use when mapping from a map[string]* to a struct or vice versa.
func JSONKeyMapper(key string, sourceTag, destTag reflect.StructTag) (string, string) {
	if s := destTag.Get("json"); len(s) > 0 {
		return strings.SplitN(s, ",", 2)[0], key
	}
	if s := sourceTag.Get("json"); len(s) > 0 {
		return key, strings.SplitN(s, ",", 2)[0]
	}
	return key, key
}

// DefaultStringConversions are helpers for converting []string and string to real values.
var DefaultStringConversions = []interface{}{
	Convert_Slice_string_To_string,
	Convert_Slice_string_To_int,
	Convert_Slice_string_To_bool,
	Convert_Slice_string_To_int64,
}

func Convert_Slice_string_To_string(input *[]string, out *string, s conversion.Scope) error {
	if len(*input) == 0 {
		*out = ""
	}
	*out = (*input)[0]
	return nil
}

func Convert_Slice_string_To_int(input *[]string, out *int, s conversion.Scope) error {
	if len(*input) == 0 {
		*out = 0
	}
	str := (*input)[0]
	i, err := strconv.Atoi(str)
	if err != nil {
		return err
	}
	*out = i
	return nil
}

// Convert_Slice_string_To_bool will convert a string parameter to boolean.
// Only the absence of a value, a value of "false", or a value of "0" resolve to false.
// Any other value (including empty string) resolves to true.
func Convert_Slice_string_To_bool(input *[]string, out *bool, s conversion.Scope) error {
	if len(*input) == 0 {
		*out = false
		return nil
	}
	switch strings.ToLower((*input)[0]) {
	case "false", "0":
		*out = false
	default:
		*out = true
	}
	return nil
}

func Convert_Slice_string_To_int64(input *[]string, out *int64, s conversion.Scope) error {
	if len(*input) == 0 {
		*out = 0
	}
	str := (*input)[0]
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return err
	}
	*out = i
	return nil
}
