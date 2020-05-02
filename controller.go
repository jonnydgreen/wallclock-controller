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

package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	projectjudgev1alpha1 "projectjudge/projects/wallclock-controller/pkg/apis/projectjudge/v1alpha1"
	clientset "projectjudge/projects/wallclock-controller/pkg/generated/clientset/versioned"
	samplescheme "projectjudge/projects/wallclock-controller/pkg/generated/clientset/versioned/scheme"
	informers "projectjudge/projects/wallclock-controller/pkg/generated/informers/externalversions/projectjudge/v1alpha1"
	listers "projectjudge/projects/wallclock-controller/pkg/generated/listers/projectjudge/v1alpha1"
	"projectjudge/projects/wallclock-controller/pkg/utils"
)

const controllerAgentName = "wallclock-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Timezone is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Timezone fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Timezone"
	// MessageResourceSynced is the message used for an Event fired when a Timezone
	// is synced successfully
	MessageResourceSynced = "Timezone synced successfully"
)

// Controller is the controller implementation for Timezone resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	sampleclientset clientset.Interface

	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced
	timezonesLister   listers.TimezoneLister
	wallClocksLister  listers.WallClockLister
	timezonesSynced   cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	sampleclientset clientset.Interface,
	deploymentInformer appsinformers.DeploymentInformer,
	timezoneInformer informers.TimezoneInformer,
	wallclockInformer informers.WallClockInformer) *Controller {

	// Create event broadcaster
	// Add wallclock-controller types to the default Kubernetes Scheme so Events can be
	// logged for wallclock-controller types.
	utilruntime.Must(samplescheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:     kubeclientset,
		sampleclientset:   sampleclientset,
		deploymentsLister: deploymentInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,
		timezonesLister:   timezoneInformer.Lister(),
		timezonesSynced:   timezoneInformer.Informer().HasSynced,
		wallClocksLister:  wallclockInformer.Lister(),
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Timezones"),
		recorder:          recorder,
	}

	klog.Info("Setting up event handlers")
	// Set up an event handler for when Timezone resources change
	timezoneInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueTimezone,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueTimezone(new)
		},
		// DeleteFunc: controller.handleObject,
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Timezone controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.deploymentsSynced, c.timezonesSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// // Launch two workers to process Timezone resources
	// for i := 0; i < threadiness; i++ {
	// }
	go wait.Until(c.runTimezoneWorker, time.Second, stopCh)

	// Launch one worker to process WallClock resources
	go wait.Until(c.runWallClockWorker, time.Second, stopCh)

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

// runTimezoneWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runTimezoneWorker() {
	for c.processNextWorkItem() {
	}
}

// runWallClockWorker is a long-running function that will run every second to update all WallClock resources with the current time
func (c *Controller) runWallClockWorker() {
	err := c.updateWallClockTimes()
	if err != nil {
		klog.Errorf("Error processing wallclocks: %v", err)
		return
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Timezone resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Timezone resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Timezone resource with this namespace/name
	timezone, err := c.timezonesLister.Get(name)
	if err != nil {
		// The Timezone resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("timezone '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	timezones := timezone.Spec.Timezones
	if len(timezones) == 0 {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		utilruntime.HandleError(fmt.Errorf("%s: timezones must be of length greater than 0", key))
		return nil
	}

	// Get WallClocks by label
	wallClockRequirement, err := labels.NewRequirement("timezoneName", selection.Equals, []string{timezone.Name})
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Unexpected error generating label selector: %v", err))
		return nil
	}
	wallClockSelector := labels.NewSelector().Add(*wallClockRequirement)
	wallClocks, err := c.wallClocksLister.List(wallClockSelector)
	if err != nil {
		return err
	}

	wallClocksToCreate := c.getWallClocksToCreate(wallClocks, timezone)
	wallClocksToDelete := c.getWallClocksToDelete(wallClocks, timezone)

	// Delete redundant WallClocks
	for _, wallClockToDelete := range wallClocksToDelete {
		err := c.sampleclientset.ProjectjudgeV1alpha1().WallClocks().Delete(context.TODO(), wallClockToDelete.Name, metav1.DeleteOptions{})
		// If an error occurs during Delete, we'll requeue the item so we can
		// attempt processing again later. This could have been caused by a
		// temporary network failure, or any other transient reason.
		if err != nil {
			return err
		}
	}

	// Create new WallClocks
	for _, wallClockToCreate := range wallClocksToCreate {
		_, err := c.sampleclientset.ProjectjudgeV1alpha1().WallClocks().Create(context.TODO(), wallClockToCreate, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	// // If the Deployment is not controlled by this Timezone resource, we should log
	// // a warning to the event recorder and return error msg.
	// if !metav1.IsControlledBy(deployment, timezone) {
	// 	msg := fmt.Sprintf(MessageResourceExists, deployment.Name)
	// 	c.recorder.Event(timezone, corev1.EventTypeWarning, ErrResourceExists, msg)
	// 	return fmt.Errorf(msg)
	// }

	// // If this number of the replicas on the Timezone resource is specified, and the
	// // number does not equal the current desired replicas on the Deployment, we
	// // should update the Deployment resource.
	// if timezone.Spec.Replicas != nil && *timezone.Spec.Replicas != *deployment.Spec.Replicas {
	// 	klog.V(4).Infof("Timezone %s replicas: %d, deployment replicas: %d", name, *timezone.Spec.Replicas, *deployment.Spec.Replicas)
	// 	deployment, err = c.kubeclientset.AppsV1().Deployments(timezone.Namespace).Update(context.TODO(), newDeployment(timezone), metav1.UpdateOptions{})
	// }

	// // If an error occurs during Update, we'll requeue the item so we can
	// // attempt processing again later. This could have been caused by a
	// // temporary network failure, or any other transient reason.
	// if err != nil {
	// 	return err
	// }

	// // Finally, we update the status block of the Timezone resource to reflect the
	// // current state of the world
	// err = c.updateTimezoneStatus(timezone, deployment)
	// if err != nil {
	// 	return err
	// }

	c.recorder.Event(timezone, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) updateWallClockTimes() error {
	// TODO: get from a store

	wallClocks, err := c.wallClocksLister.List(labels.NewSelector())
	if err != nil {
		return err
	}

	// Update WallClocks with the latest time
	for _, wallClock := range wallClocks {
		c.updateWallClockTime(wallClock.DeepCopy())
	}
	return err
}

func (c *Controller) updateWallClockTime(wallClock *projectjudgev1alpha1.WallClock) error {
	loc, err := time.LoadLocation(wallClock.Spec.Timezone)
	if err != nil {
		klog.Errorf("Error loading timezone: %v", err)
	} else {
		localTime := time.Now().In(loc).Format("15:04:05")
		wallClock.Status.Time = localTime
		_, err = c.sampleclientset.ProjectjudgeV1alpha1().WallClocks().UpdateStatus(context.TODO(), wallClock, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorf("Error updating wallclock: %v", err)
		}
	}
	return err
}

// enqueueTimezone takes a Timezone resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Timezone.
func (c *Controller) enqueueTimezone(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *Controller) getWallClocksToCreate(wallClocks []*projectjudgev1alpha1.WallClock, timezone *projectjudgev1alpha1.Timezone) []*projectjudgev1alpha1.WallClock {
	wallClocksToCreate := []*projectjudgev1alpha1.WallClock{}
TimezoneLabel:
	for _, timezoneString := range timezone.Spec.Timezones {
		for _, wallClock := range wallClocks {
			if timezoneString == (*wallClock).Spec.Timezone {
				continue TimezoneLabel
			}
		}
		wallClocksToCreate = append(wallClocksToCreate, newWallClock(timezone, timezoneString))
	}
	return wallClocksToCreate
}

func (c *Controller) getWallClocksToDelete(wallClocks []*projectjudgev1alpha1.WallClock, timezone *projectjudgev1alpha1.Timezone) []*projectjudgev1alpha1.WallClock {
	wallClocksToDelete := []*projectjudgev1alpha1.WallClock{}
	for _, wallClock := range wallClocks {
		if !utils.Contains(timezone.Spec.Timezones, (*wallClock).Spec.Timezone) {
			wallClocksToDelete = append(wallClocksToDelete, wallClock)
		}
	}
	return wallClocksToDelete
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the Timezone resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that Timezone resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		klog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	klog.V(4).Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a Timezone, we should not do anything more
		// with it.
		if ownerRef.Kind != "Timezone" {
			return
		}

		timezone, err := c.timezonesLister.Get(ownerRef.Name)
		if err != nil {
			klog.V(4).Infof("ignoring orphaned object '%s' of timezone '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueueTimezone(timezone)
		return
	}
}

func constructWallClockName(timezoneName string, timezone string) string {
	formattedTimezone := strings.ReplaceAll(timezoneName+"-"+timezone, "/", "-")
	formattedTimezone = strings.ReplaceAll(formattedTimezone, "_", "-")
	return strings.ToLower(formattedTimezone)
}

// newWallClock creates a new WallClock for a Timezone resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Timezone resource that 'owns' it.
func newWallClock(timezone *projectjudgev1alpha1.Timezone, timezoneString string) *projectjudgev1alpha1.WallClock {
	labels := map[string]string{
		"timezoneName": timezone.Name,
		"controller":   timezone.Name,
	}
	return &projectjudgev1alpha1.WallClock{
		ObjectMeta: metav1.ObjectMeta{
			Name: constructWallClockName(timezone.Name, timezoneString),
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(timezone, projectjudgev1alpha1.SchemeGroupVersion.WithKind("Timezone")),
			},
			Labels: labels,
		},
		Spec: projectjudgev1alpha1.WallClockSpec{
			Timezone: timezoneString,
		},
	}
}
