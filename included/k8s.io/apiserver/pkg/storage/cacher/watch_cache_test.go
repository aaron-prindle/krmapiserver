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

package cacher

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/api/core/v1"
	apiequality "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/api/equality"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/api/errors"
	metav1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/fields"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/labels"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/clock"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/wait"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/watch"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/storage"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/storage/etcd"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/tools/cache"
)

func makeTestPod(name string, resourceVersion uint64) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       "ns",
			Name:            name,
			ResourceVersion: strconv.FormatUint(resourceVersion, 10),
			Labels: map[string]string{
				"k8s-app": "my-app",
			},
		},
		Spec: v1.PodSpec{
			NodeName: "some-node",
		},
	}
}

func makeTestStoreElement(pod *v1.Pod) *storeElement {
	return &storeElement{
		Key:    "prefix/ns/" + pod.Name,
		Object: pod,
		Labels: labels.Set(pod.Labels),
		Fields: fields.Set{"spec.nodeName": pod.Spec.NodeName},
	}
}

// newTestWatchCache just adds a fake clock.
func newTestWatchCache(capacity int) *watchCache {
	keyFunc := func(obj runtime.Object) (string, error) {
		return storage.NamespaceKeyFunc("prefix", obj)
	}
	getAttrsFunc := func(obj runtime.Object) (labels.Set, fields.Set, error) {
		pod, ok := obj.(*v1.Pod)
		if !ok {
			return nil, nil, fmt.Errorf("not a pod")
		}
		return labels.Set(pod.Labels), fields.Set{"spec.nodeName": pod.Spec.NodeName}, nil
	}
	versioner := etcd.APIObjectVersioner{}
	mockHandler := func(*watchCacheEvent) {}
	wc := newWatchCache(capacity, keyFunc, mockHandler, getAttrsFunc, versioner)
	wc.clock = clock.NewFakeClock(time.Now())
	return wc
}

func TestWatchCacheBasic(t *testing.T) {
	store := newTestWatchCache(2)

	// Test Add/Update/Delete.
	pod1 := makeTestPod("pod", 1)
	if err := store.Add(pod1); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if item, ok, _ := store.Get(pod1); !ok {
		t.Errorf("didn't find pod")
	} else {
		expected := makeTestStoreElement(makeTestPod("pod", 1))
		if !apiequality.Semantic.DeepEqual(expected, item) {
			t.Errorf("expected %v, got %v", expected, item)
		}
	}
	pod2 := makeTestPod("pod", 2)
	if err := store.Update(pod2); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if item, ok, _ := store.Get(pod2); !ok {
		t.Errorf("didn't find pod")
	} else {
		expected := makeTestStoreElement(makeTestPod("pod", 2))
		if !apiequality.Semantic.DeepEqual(expected, item) {
			t.Errorf("expected %v, got %v", expected, item)
		}
	}
	pod3 := makeTestPod("pod", 3)
	if err := store.Delete(pod3); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if _, ok, _ := store.Get(pod3); ok {
		t.Errorf("found pod")
	}

	// Test List.
	store.Add(makeTestPod("pod1", 4))
	store.Add(makeTestPod("pod2", 5))
	store.Add(makeTestPod("pod3", 6))
	{
		expected := map[string]storeElement{
			"prefix/ns/pod1": *makeTestStoreElement(makeTestPod("pod1", 4)),
			"prefix/ns/pod2": *makeTestStoreElement(makeTestPod("pod2", 5)),
			"prefix/ns/pod3": *makeTestStoreElement(makeTestPod("pod3", 6)),
		}
		items := make(map[string]storeElement, 0)
		for _, item := range store.List() {
			elem := item.(*storeElement)
			items[elem.Key] = *elem
		}
		if !apiequality.Semantic.DeepEqual(expected, items) {
			t.Errorf("expected %v, got %v", expected, items)
		}
	}

	// Test Replace.
	store.Replace([]interface{}{
		makeTestPod("pod4", 7),
		makeTestPod("pod5", 8),
	}, "8")
	{
		expected := map[string]storeElement{
			"prefix/ns/pod4": *makeTestStoreElement(makeTestPod("pod4", 7)),
			"prefix/ns/pod5": *makeTestStoreElement(makeTestPod("pod5", 8)),
		}
		items := make(map[string]storeElement)
		for _, item := range store.List() {
			elem := item.(*storeElement)
			items[elem.Key] = *elem
		}
		if !apiequality.Semantic.DeepEqual(expected, items) {
			t.Errorf("expected %v, got %v", expected, items)
		}
	}
}

func TestEvents(t *testing.T) {
	store := newTestWatchCache(5)

	store.Add(makeTestPod("pod", 3))

	// Test for Added event.
	{
		_, err := store.GetAllEventsSince(1)
		if err == nil {
			t.Errorf("expected error too old")
		}
		if _, ok := err.(*errors.StatusError); !ok {
			t.Errorf("expected error to be of type StatusError")
		}
	}
	{
		result, err := store.GetAllEventsSince(2)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("unexpected events: %v", result)
		}
		if result[0].Type != watch.Added {
			t.Errorf("unexpected event type: %v", result[0].Type)
		}
		pod := makeTestPod("pod", uint64(3))
		if !apiequality.Semantic.DeepEqual(pod, result[0].Object) {
			t.Errorf("unexpected item: %v, expected: %v", result[0].Object, pod)
		}
		if result[0].PrevObject != nil {
			t.Errorf("unexpected item: %v", result[0].PrevObject)
		}
	}

	store.Update(makeTestPod("pod", 4))
	store.Update(makeTestPod("pod", 5))

	// Test with not full cache.
	{
		_, err := store.GetAllEventsSince(1)
		if err == nil {
			t.Errorf("expected error too old")
		}
	}
	{
		result, err := store.GetAllEventsSince(3)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(result) != 2 {
			t.Fatalf("unexpected events: %v", result)
		}
		for i := 0; i < 2; i++ {
			if result[i].Type != watch.Modified {
				t.Errorf("unexpected event type: %v", result[i].Type)
			}
			pod := makeTestPod("pod", uint64(i+4))
			if !apiequality.Semantic.DeepEqual(pod, result[i].Object) {
				t.Errorf("unexpected item: %v, expected: %v", result[i].Object, pod)
			}
			prevPod := makeTestPod("pod", uint64(i+3))
			if !apiequality.Semantic.DeepEqual(prevPod, result[i].PrevObject) {
				t.Errorf("unexpected item: %v, expected: %v", result[i].PrevObject, prevPod)
			}
		}
	}

	for i := 6; i < 10; i++ {
		store.Update(makeTestPod("pod", uint64(i)))
	}

	// Test with full cache - there should be elements from 5 to 9.
	{
		_, err := store.GetAllEventsSince(3)
		if err == nil {
			t.Errorf("expected error too old")
		}
	}
	{
		result, err := store.GetAllEventsSince(4)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(result) != 5 {
			t.Fatalf("unexpected events: %v", result)
		}
		for i := 0; i < 5; i++ {
			pod := makeTestPod("pod", uint64(i+5))
			if !apiequality.Semantic.DeepEqual(pod, result[i].Object) {
				t.Errorf("unexpected item: %v, expected: %v", result[i].Object, pod)
			}
		}
	}

	// Test for delete event.
	store.Delete(makeTestPod("pod", uint64(10)))

	{
		result, err := store.GetAllEventsSince(9)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("unexpected events: %v", result)
		}
		if result[0].Type != watch.Deleted {
			t.Errorf("unexpected event type: %v", result[0].Type)
		}
		pod := makeTestPod("pod", uint64(10))
		if !apiequality.Semantic.DeepEqual(pod, result[0].Object) {
			t.Errorf("unexpected item: %v, expected: %v", result[0].Object, pod)
		}
		prevPod := makeTestPod("pod", uint64(9))
		if !apiequality.Semantic.DeepEqual(prevPod, result[0].PrevObject) {
			t.Errorf("unexpected item: %v, expected: %v", result[0].PrevObject, prevPod)
		}
	}
}

func TestMarker(t *testing.T) {
	store := newTestWatchCache(3)

	// First thing that is called when propagated from storage is Replace.
	store.Replace([]interface{}{
		makeTestPod("pod1", 5),
		makeTestPod("pod2", 9),
	}, "9")

	_, err := store.GetAllEventsSince(8)
	if err == nil || !strings.Contains(err.Error(), "too old resource version") {
		t.Errorf("unexpected error: %v", err)
	}
	// Getting events from 8 should return no events,
	// even though there is a marker there.
	result, err := store.GetAllEventsSince(9)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("unexpected result: %#v, expected no events", result)
	}

	pod := makeTestPod("pods", 12)
	store.Add(pod)
	// Getting events from 8 should still work and return one event.
	result, err = store.GetAllEventsSince(9)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || !apiequality.Semantic.DeepEqual(result[0].Object, pod) {
		t.Errorf("unexpected result: %#v, expected %v", result, pod)
	}
}

func TestWaitUntilFreshAndList(t *testing.T) {
	store := newTestWatchCache(3)

	// In background, update the store.
	go func() {
		store.Add(makeTestPod("foo", 2))
		store.Add(makeTestPod("bar", 5))
	}()

	list, resourceVersion, err := store.WaitUntilFreshAndList(5, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resourceVersion != 5 {
		t.Errorf("unexpected resourceVersion: %v, expected: 5", resourceVersion)
	}
	if len(list) != 2 {
		t.Errorf("unexpected list returned: %#v", list)
	}
}

func TestWaitUntilFreshAndGet(t *testing.T) {
	store := newTestWatchCache(3)

	// In background, update the store.
	go func() {
		store.Add(makeTestPod("foo", 2))
		store.Add(makeTestPod("bar", 5))
	}()

	obj, exists, resourceVersion, err := store.WaitUntilFreshAndGet(5, "prefix/ns/bar", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resourceVersion != 5 {
		t.Errorf("unexpected resourceVersion: %v, expected: 5", resourceVersion)
	}
	if !exists {
		t.Fatalf("no results returned: %#v", obj)
	}
	expected := makeTestStoreElement(makeTestPod("bar", 5))
	if !apiequality.Semantic.DeepEqual(expected, obj) {
		t.Errorf("expected %v, got %v", expected, obj)
	}
}

func TestWaitUntilFreshAndListTimeout(t *testing.T) {
	store := newTestWatchCache(3)
	fc := store.clock.(*clock.FakeClock)

	// In background, step clock after the below call starts the timer.
	go func() {
		for !fc.HasWaiters() {
			time.Sleep(time.Millisecond)
		}
		fc.Step(blockTimeout)

		// Add an object to make sure the test would
		// eventually fail instead of just waiting
		// forever.
		time.Sleep(30 * time.Second)
		store.Add(makeTestPod("bar", 5))
	}()

	_, _, err := store.WaitUntilFreshAndList(5, nil)
	if err == nil {
		t.Fatalf("unexpected lack of timeout error")
	}
}

type testLW struct {
	ListFunc  func(options metav1.ListOptions) (runtime.Object, error)
	WatchFunc func(options metav1.ListOptions) (watch.Interface, error)
}

func (t *testLW) List(options metav1.ListOptions) (runtime.Object, error) {
	return t.ListFunc(options)
}
func (t *testLW) Watch(options metav1.ListOptions) (watch.Interface, error) {
	return t.WatchFunc(options)
}

func TestReflectorForWatchCache(t *testing.T) {
	store := newTestWatchCache(5)

	{
		_, version, err := store.WaitUntilFreshAndList(0, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if version != 0 {
			t.Errorf("unexpected resource version: %d", version)
		}
	}

	lw := &testLW{
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			fw := watch.NewFake()
			go fw.Stop()
			return fw, nil
		},
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return &v1.PodList{ListMeta: metav1.ListMeta{ResourceVersion: "10"}}, nil
		},
	}
	r := cache.NewReflector(lw, &v1.Pod{}, store, 0)
	r.ListAndWatch(wait.NeverStop)

	{
		_, version, err := store.WaitUntilFreshAndList(10, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if version != 10 {
			t.Errorf("unexpected resource version: %d", version)
		}
	}
}
