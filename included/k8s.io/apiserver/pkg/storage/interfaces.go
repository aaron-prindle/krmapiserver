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

package storage

import (
	"context"
	"fmt"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/api/meta"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/fields"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/labels"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/types"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/watch"
)

// Versioner abstracts setting and retrieving metadata fields from database response
// onto the object ot list. It is required to maintain storage invariants - updating an
// object twice with the same data except for the ResourceVersion and SelfLink must be
// a no-op. A resourceVersion of type uint64 is a 'raw' resourceVersion,
// intended to be sent directly to or from the backend. A resourceVersion of
// type string is a 'safe' resourceVersion, intended for consumption by users.
type Versioner interface {
	// UpdateObject sets storage metadata into an API object. Returns an error if the object
	// cannot be updated correctly. May return nil if the requested object does not need metadata
	// from database.
	UpdateObject(obj runtime.Object, resourceVersion uint64) error
	// UpdateList sets the resource version into an API list object. Returns an error if the object
	// cannot be updated correctly. May return nil if the requested object does not need metadata from
	// database. continueValue is optional and indicates that more results are available if the client
	// passes that value to the server in a subsequent call. remainingItemCount indicates the number
	// of remaining objects if the list is partial. The remainingItemCount field is omitted during
	// serialization if it is set to nil.
	UpdateList(obj runtime.Object, resourceVersion uint64, continueValue string, remainingItemCount *int64) error
	// PrepareObjectForStorage should set SelfLink and ResourceVersion to the empty value. Should
	// return an error if the specified object cannot be updated.
	PrepareObjectForStorage(obj runtime.Object) error
	// ObjectResourceVersion returns the resource version (for persistence) of the specified object.
	// Should return an error if the specified object does not have a persistable version.
	ObjectResourceVersion(obj runtime.Object) (uint64, error)

	// ParseResourceVersion takes a resource version argument and
	// converts it to the storage backend. For watch we should pass to helper.Watch().
	// Because resourceVersion is an opaque value, the default watch
	// behavior for non-zero watch is to watch the next value (if you pass
	// "1", you will see updates from "2" onwards).
	ParseResourceVersion(resourceVersion string) (uint64, error)
}

// ResponseMeta contains information about the database metadata that is associated with
// an object. It abstracts the actual underlying objects to prevent coupling with concrete
// database and to improve testability.
type ResponseMeta struct {
	// TTL is the time to live of the node that contained the returned object. It may be
	// zero or negative in some cases (objects may be expired after the requested
	// expiration time due to server lag).
	TTL int64
	// The resource version of the node that contained the returned object.
	ResourceVersion uint64
}

// MatchValue defines a pair (<index name>, <value for that index>).
type MatchValue struct {
	IndexName string
	Value     string
}

// TriggerPublisherFunc is a function that takes an object, and returns a list of pairs
// (<index name>, <index value for the given object>) for all indexes known
// to that function.
type TriggerPublisherFunc func(obj runtime.Object) []MatchValue

// Everything accepts all objects.
var Everything = SelectionPredicate{
	Label: labels.Everything(),
	Field: fields.Everything(),
}

// Pass an UpdateFunc to Interface.GuaranteedUpdate to make an update
// that is guaranteed to succeed.
// See the comment for GuaranteedUpdate for more details.
type UpdateFunc func(input runtime.Object, res ResponseMeta) (output runtime.Object, ttl *uint64, err error)

// ValidateObjectFunc is a function to act on a given object. An error may be returned
// if the hook cannot be completed. The function may NOT transform the provided
// object.
type ValidateObjectFunc func(obj runtime.Object) error

// ValidateAllObjectFunc is a "admit everything" instance of ValidateObjectFunc.
func ValidateAllObjectFunc(obj runtime.Object) error {
	return nil
}

// Preconditions must be fulfilled before an operation (update, delete, etc.) is carried out.
type Preconditions struct {
	// Specifies the target UID.
	// +optional
	UID *types.UID `json:"uid,omitempty"`
	// Specifies the target ResourceVersion
	// +optional
	ResourceVersion *string `json:"resourceVersion,omitempty"`
}

// NewUIDPreconditions returns a Preconditions with UID set.
func NewUIDPreconditions(uid string) *Preconditions {
	u := types.UID(uid)
	return &Preconditions{UID: &u}
}

func (p *Preconditions) Check(key string, obj runtime.Object) error {
	if p == nil {
		return nil
	}
	objMeta, err := meta.Accessor(obj)
	if err != nil {
		return NewInternalErrorf(
			"can't enforce preconditions %v on un-introspectable object %v, got error: %v",
			*p,
			obj,
			err)
	}
	if p.UID != nil && *p.UID != objMeta.GetUID() {
		err := fmt.Sprintf(
			"Precondition failed: UID in precondition: %v, UID in object meta: %v",
			*p.UID,
			objMeta.GetUID())
		return NewInvalidObjError(key, err)
	}
	if p.ResourceVersion != nil && *p.ResourceVersion != objMeta.GetResourceVersion() {
		err := fmt.Sprintf(
			"Precondition failed: ResourceVersion in precondition: %v, ResourceVersion in object meta: %v",
			*p.ResourceVersion,
			objMeta.GetResourceVersion())
		return NewInvalidObjError(key, err)
	}
	return nil
}

// Interface offers a common interface for object marshaling/unmarshaling operations and
// hides all the storage-related operations behind it.
type Interface interface {
	// Returns Versioner associated with this interface.
	Versioner() Versioner

	// Create adds a new object at a key unless it already exists. 'ttl' is time-to-live
	// in seconds (0 means forever). If no error is returned and out is not nil, out will be
	// set to the read value from database.
	Create(ctx context.Context, key string, obj, out runtime.Object, ttl uint64) error

	// Delete removes the specified key and returns the value that existed at that spot.
	// If key didn't exist, it will return NotFound storage error.
	Delete(ctx context.Context, key string, out runtime.Object, preconditions *Preconditions, validateDeletion ValidateObjectFunc) error

	// Watch begins watching the specified key. Events are decoded into API objects,
	// and any items selected by 'p' are sent down to returned watch.Interface.
	// resourceVersion may be used to specify what version to begin watching,
	// which should be the current resourceVersion, and no longer rv+1
	// (e.g. reconnecting without missing any updates).
	// If resource version is "0", this interface will get current object at given key
	// and send it in an "ADDED" event, before watch starts.
	Watch(ctx context.Context, key string, resourceVersion string, p SelectionPredicate) (watch.Interface, error)

	// WatchList begins watching the specified key's items. Items are decoded into API
	// objects and any item selected by 'p' are sent down to returned watch.Interface.
	// resourceVersion may be used to specify what version to begin watching,
	// which should be the current resourceVersion, and no longer rv+1
	// (e.g. reconnecting without missing any updates).
	// If resource version is "0", this interface will list current objects directory defined by key
	// and send them in "ADDED" events, before watch starts.
	WatchList(ctx context.Context, key string, resourceVersion string, p SelectionPredicate) (watch.Interface, error)

	// Get unmarshals json found at key into objPtr. On a not found error, will either
	// return a zero object of the requested type, or an error, depending on ignoreNotFound.
	// Treats empty responses and nil response nodes exactly like a not found error.
	// The returned contents may be delayed, but it is guaranteed that they will
	// be have at least 'resourceVersion'.
	Get(ctx context.Context, key string, resourceVersion string, objPtr runtime.Object, ignoreNotFound bool) error

	// GetToList unmarshals json found at key and opaque it into *List api object
	// (an object that satisfies the runtime.IsList definition).
	// The returned contents may be delayed, but it is guaranteed that they will
	// be have at least 'resourceVersion'.
	GetToList(ctx context.Context, key string, resourceVersion string, p SelectionPredicate, listObj runtime.Object) error

	// List unmarshalls jsons found at directory defined by key and opaque them
	// into *List api object (an object that satisfies runtime.IsList definition).
	// The returned contents may be delayed, but it is guaranteed that they will
	// be have at least 'resourceVersion'.
	List(ctx context.Context, key string, resourceVersion string, p SelectionPredicate, listObj runtime.Object) error

	// GuaranteedUpdate keeps calling 'tryUpdate()' to update key 'key' (of type 'ptrToType')
	// retrying the update until success if there is index conflict.
	// Note that object passed to tryUpdate may change across invocations of tryUpdate() if
	// other writers are simultaneously updating it, so tryUpdate() needs to take into account
	// the current contents of the object when deciding how the update object should look.
	// If the key doesn't exist, it will return NotFound storage error if ignoreNotFound=false
	// or zero value in 'ptrToType' parameter otherwise.
	// If the object to update has the same value as previous, it won't do any update
	// but will return the object in 'ptrToType' parameter.
	// If 'suggestion' can contain zero or one element - in such case this can be used as
	// a suggestion about the current version of the object to avoid read operation from
	// storage to get it.
	//
	// Example:
	//
	// s := /* implementation of Interface */
	// err := s.GuaranteedUpdate(
	//     "myKey", &MyType{}, true,
	//     func(input runtime.Object, res ResponseMeta) (runtime.Object, *uint64, error) {
	//       // Before each incovation of the user defined function, "input" is reset to
	//       // current contents for "myKey" in database.
	//       curr := input.(*MyType)  // Guaranteed to succeed.
	//
	//       // Make the modification
	//       curr.Counter++
	//
	//       // Return the modified object - return an error to stop iterating. Return
	//       // a uint64 to alter the TTL on the object, or nil to keep it the same value.
	//       return cur, nil, nil
	//    },
	// )
	GuaranteedUpdate(
		ctx context.Context, key string, ptrToType runtime.Object, ignoreNotFound bool,
		precondtions *Preconditions, tryUpdate UpdateFunc, suggestion ...runtime.Object) error

	// Count returns number of different entries under the key (generally being path prefix).
	Count(key string) (int64, error)
}
