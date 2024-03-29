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

package value

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aaron-prindle/krmapiserver/included/google.golang.org/grpc/codes"
	"github.com/aaron-prindle/krmapiserver/included/google.golang.org/grpc/status"

	"github.com/aaron-prindle/krmapiserver/included/github.com/prometheus/client_golang/prometheus"
	"github.com/aaron-prindle/krmapiserver/included/github.com/prometheus/client_golang/prometheus/testutil"
)

func TestTotals(t *testing.T) {
	testCases := []struct {
		desc    string
		metrics []string
		error   error
		want    string
	}{
		{
			desc: "non-status error",
			metrics: []string{
				"apiserver_storage_transformation_operations_total",
				"apiserver_storage_transformation_failures_total",
			},
			error: errors.New("foo"),
			want: `
  # HELP apiserver_storage_transformation_failures_total (Deprecated) Total number of failed transformation operations.
  # TYPE apiserver_storage_transformation_failures_total counter
  apiserver_storage_transformation_failures_total{transformation_type="encrypt"} 1
	# HELP apiserver_storage_transformation_operations_total Total number of transformations.
  # TYPE apiserver_storage_transformation_operations_total counter
  apiserver_storage_transformation_operations_total{status="Unknown",transformation_type="encrypt"} 1
`,
		},
		{
			desc: "error is nil",
			metrics: []string{
				"apiserver_storage_transformation_operations_total",
				"apiserver_storage_transformation_failures_total",
			},
			want: `
	# HELP apiserver_storage_transformation_operations_total Total number of transformations.
  # TYPE apiserver_storage_transformation_operations_total counter
  apiserver_storage_transformation_operations_total{status="OK",transformation_type="encrypt"} 1
`,
		},
		{
			desc: "status error from kms-plugin",
			metrics: []string{
				"apiserver_storage_transformation_operations_total",
				"apiserver_storage_transformation_failures_total",
			},
			error: status.Error(codes.FailedPrecondition, "foo"),
			want: `
  # HELP apiserver_storage_transformation_failures_total (Deprecated) Total number of failed transformation operations.
  # TYPE apiserver_storage_transformation_failures_total counter
  apiserver_storage_transformation_failures_total{transformation_type="encrypt"} 1
	# HELP apiserver_storage_transformation_operations_total Total number of transformations.
  # TYPE apiserver_storage_transformation_operations_total counter
  apiserver_storage_transformation_operations_total{status="FailedPrecondition",transformation_type="encrypt"} 1
`,
		},
	}

	RegisterMetrics()

	for _, tt := range testCases {
		t.Run(tt.desc, func(t *testing.T) {
			RecordTransformation("encrypt", time.Now(), tt.error)
			defer transformerOperationsTotal.Reset()
			defer deprecatedTransformerFailuresTotal.Reset()
			if err := testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(tt.want), tt.metrics...); err != nil {
				t.Fatal(err)
			}
		})
	}
}
