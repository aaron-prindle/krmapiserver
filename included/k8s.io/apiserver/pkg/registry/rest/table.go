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

package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/api/meta"
	metav1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime/schema"
)

type defaultTableConvertor struct {
	qualifiedResource schema.GroupResource
}

// NewDefaultTableConvertor creates a default convertor for the provided resource.
func NewDefaultTableConvertor(resource schema.GroupResource) TableConvertor {
	return defaultTableConvertor{qualifiedResource: resource}
}

var swaggerMetadataDescriptions = metav1.ObjectMeta{}.SwaggerDoc()

func (c defaultTableConvertor) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1beta1.Table, error) {
	var table metav1beta1.Table
	fn := func(obj runtime.Object) error {
		m, err := meta.Accessor(obj)
		if err != nil {
			return errNotAcceptable{resource: c.qualifiedResource}
		}
		table.Rows = append(table.Rows, metav1beta1.TableRow{
			Cells:  []interface{}{m.GetName(), m.GetCreationTimestamp().Time.UTC().Format(time.RFC3339)},
			Object: runtime.RawExtension{Object: obj},
		})
		return nil
	}
	switch {
	case meta.IsListType(object):
		if err := meta.EachListItem(object, fn); err != nil {
			return nil, err
		}
	default:
		if err := fn(object); err != nil {
			return nil, err
		}
	}
	if m, err := meta.ListAccessor(object); err == nil {
		table.ResourceVersion = m.GetResourceVersion()
		table.SelfLink = m.GetSelfLink()
		table.Continue = m.GetContinue()
		table.RemainingItemCount = m.GetRemainingItemCount()
	} else {
		if m, err := meta.CommonAccessor(object); err == nil {
			table.ResourceVersion = m.GetResourceVersion()
			table.SelfLink = m.GetSelfLink()
		}
	}
	if opt, ok := tableOptions.(*metav1beta1.TableOptions); !ok || !opt.NoHeaders {
		table.ColumnDefinitions = []metav1beta1.TableColumnDefinition{
			{Name: "Name", Type: "string", Format: "name", Description: swaggerMetadataDescriptions["name"]},
			{Name: "Created At", Type: "date", Description: swaggerMetadataDescriptions["creationTimestamp"]},
		}
	}
	return &table, nil
}

// errNotAcceptable indicates the resource doesn't support Table conversion
type errNotAcceptable struct {
	resource schema.GroupResource
}

func (e errNotAcceptable) Error() string {
	return fmt.Sprintf("the resource %s does not support being converted to a Table", e.resource)
}

func (e errNotAcceptable) Status() metav1.Status {
	return metav1.Status{
		Status:  metav1.StatusFailure,
		Code:    http.StatusNotAcceptable,
		Reason:  metav1.StatusReason("NotAcceptable"),
		Message: e.Error(),
	}
}
