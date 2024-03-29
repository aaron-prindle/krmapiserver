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

package metrics

import (
	"fmt"
	"sync"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/util/flowcontrol"

	"github.com/aaron-prindle/krmapiserver/included/github.com/prometheus/client_golang/prometheus"
)

var (
	metricsLock        sync.Mutex
	rateLimiterMetrics = make(map[string]*rateLimiterMetric)
)

type rateLimiterMetric struct {
	metric prometheus.Gauge
	stopCh chan struct{}
}

func registerRateLimiterMetric(ownerName string) error {
	metricsLock.Lock()
	defer metricsLock.Unlock()

	if _, ok := rateLimiterMetrics[ownerName]; ok {
		// only register once in Prometheus. We happen to see an ownerName reused in parallel integration tests.
		return nil
	}
	metric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "rate_limiter_use",
		Subsystem: ownerName,
		Help:      fmt.Sprintf("A metric measuring the saturation of the rate limiter for %v", ownerName),
	})
	if err := prometheus.Register(metric); err != nil {
		return fmt.Errorf("error registering rate limiter usage metric: %v", err)
	}
	stopCh := make(chan struct{})
	rateLimiterMetrics[ownerName] = &rateLimiterMetric{
		metric: metric,
		stopCh: stopCh,
	}
	return nil
}

// RegisterMetricAndTrackRateLimiterUsage registers a metric ownerName_rate_limiter_use in prometheus to track
// how much used rateLimiter is and starts a goroutine that updates this metric every updatePeriod
func RegisterMetricAndTrackRateLimiterUsage(ownerName string, rateLimiter flowcontrol.RateLimiter) error {
	if err := registerRateLimiterMetric(ownerName); err != nil {
		return err
	}
	// TODO: determine how to track rate limiter saturation
	// See discussion at https://go-review.googlesource.com/c/time/+/29958#message-4caffc11669cadd90e2da4c05122cfec50ea6a22
	// go wait.Until(func() {
	//   metricsLock.Lock()
	//   defer metricsLock.Unlock()
	//   rateLimiterMetrics[ownerName].metric.Set()
	// }, updatePeriod, rateLimiterMetrics[ownerName].stopCh)
	return nil
}
