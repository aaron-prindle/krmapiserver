/*
Copyright 2016 The Kubernetes Authors.

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

package apiserver

import (
	"fmt"
	"time"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/klog"

	apierrors "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/api/errors"
	utilruntime "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/runtime"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/wait"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/tools/cache"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/client-go/util/workqueue"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/kube-aggregator/pkg/apis/apiregistration"
	informers "github.com/aaron-prindle/krmapiserver/included/k8s.io/kube-aggregator/pkg/client/informers/internalversion/apiregistration/internalversion"
	listers "github.com/aaron-prindle/krmapiserver/included/k8s.io/kube-aggregator/pkg/client/listers/apiregistration/internalversion"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/kube-aggregator/pkg/controllers"
)

// APIHandlerManager defines the behaviour that an API handler should have.
type APIHandlerManager interface {
	AddAPIService(apiService *apiregistration.APIService) error
	RemoveAPIService(apiServiceName string)
}

// APIServiceRegistrationController is responsible for registering and removing API services.
type APIServiceRegistrationController struct {
	apiHandlerManager APIHandlerManager

	apiServiceLister listers.APIServiceLister
	apiServiceSynced cache.InformerSynced

	// To allow injection for testing.
	syncFn func(key string) error

	queue workqueue.RateLimitingInterface
}

// NewAPIServiceRegistrationController returns a new APIServiceRegistrationController.
func NewAPIServiceRegistrationController(apiServiceInformer informers.APIServiceInformer, apiHandlerManager APIHandlerManager) *APIServiceRegistrationController {
	c := &APIServiceRegistrationController{
		apiHandlerManager: apiHandlerManager,
		apiServiceLister:  apiServiceInformer.Lister(),
		apiServiceSynced:  apiServiceInformer.Informer().HasSynced,
		queue:             workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "APIServiceRegistrationController"),
	}

	apiServiceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addAPIService,
		UpdateFunc: c.updateAPIService,
		DeleteFunc: c.deleteAPIService,
	})

	c.syncFn = c.sync

	return c
}

func (c *APIServiceRegistrationController) sync(key string) error {
	apiService, err := c.apiServiceLister.Get(key)
	if apierrors.IsNotFound(err) {
		c.apiHandlerManager.RemoveAPIService(key)
		return nil
	}
	if err != nil {
		return err
	}

	return c.apiHandlerManager.AddAPIService(apiService)
}

// Run starts APIServiceRegistrationController which will process all registration requests until stopCh is closed.
func (c *APIServiceRegistrationController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Infof("Starting APIServiceRegistrationController")
	defer klog.Infof("Shutting down APIServiceRegistrationController")

	if !controllers.WaitForCacheSync("APIServiceRegistrationController", stopCh, c.apiServiceSynced) {
		return
	}

	// only start one worker thread since its a slow moving API and the aggregation server adding bits
	// aren't threadsafe
	go wait.Until(c.runWorker, time.Second, stopCh)

	<-stopCh
}

func (c *APIServiceRegistrationController) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false when it's time to quit.
func (c *APIServiceRegistrationController) processNextWorkItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.syncFn(key.(string))
	if err == nil {
		c.queue.Forget(key)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("%v failed with : %v", key, err))
	c.queue.AddRateLimited(key)

	return true
}

func (c *APIServiceRegistrationController) enqueue(obj *apiregistration.APIService) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		klog.Errorf("Couldn't get key for object %#v: %v", obj, err)
		return
	}

	c.queue.Add(key)
}

func (c *APIServiceRegistrationController) addAPIService(obj interface{}) {
	castObj := obj.(*apiregistration.APIService)
	klog.V(4).Infof("Adding %s", castObj.Name)
	c.enqueue(castObj)
}

func (c *APIServiceRegistrationController) updateAPIService(obj, _ interface{}) {
	castObj := obj.(*apiregistration.APIService)
	klog.V(4).Infof("Updating %s", castObj.Name)
	c.enqueue(castObj)
}

func (c *APIServiceRegistrationController) deleteAPIService(obj interface{}) {
	castObj, ok := obj.(*apiregistration.APIService)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			klog.Errorf("Couldn't get object from tombstone %#v", obj)
			return
		}
		castObj, ok = tombstone.Obj.(*apiregistration.APIService)
		if !ok {
			klog.Errorf("Tombstone contained object that is not expected %#v", obj)
			return
		}
	}
	klog.V(4).Infof("Deleting %q", castObj.Name)
	c.enqueue(castObj)
}
