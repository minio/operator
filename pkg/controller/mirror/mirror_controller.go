/*
 * Minio-Operator - Manage Minio clusters in Kubernetes
 *
 * Minio (C) 2018 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mirror

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang/glog"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	scheme "k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	queue "k8s.io/client-go/util/workqueue"

	miniov1beta1 "github.com/minio/minio-operator/pkg/apis/miniocontroller/v1beta1"
	clientset "github.com/minio/minio-operator/pkg/client/clientset/versioned"
	minioscheme "github.com/minio/minio-operator/pkg/client/clientset/versioned/scheme"
	informers "github.com/minio/minio-operator/pkg/client/informers/externalversions/miniocontroller/v1beta1"
	listers "github.com/minio/minio-operator/pkg/client/listers/miniocontroller/v1beta1"
	constants "github.com/minio/minio-operator/pkg/constants"
	services "github.com/minio/minio-operator/pkg/resources/services"
	statefulsets "github.com/minio/minio-operator/pkg/resources/statefulsets"
)

const controllerAgentName = "minio-operator"

const (
	// SuccessSynced is used as part of the Event 'reason' when a MinioInstance is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a MinioInstance fails
	// to sync due to a StatefulSet of the same name already existing.
	ErrResourceExists = "ErrResourceExists"
	// MessageResourceExists is the message used for Events when a MinioInstance
	// fails to sync due to a StatefulSet already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Minio"
	// MessageResourceSynced is the message used for an Event fired when a MinioInstance
	// is synced successfully
	MessageResourceSynced = "MinioInstance synced successfully"
)

// Controller struct watches the Kubernetes API for changes to MinioInstance resources
type Controller struct {
	// kubeClientSet is a standard kubernetes clientset
	kubeClientSet kubernetes.Interface
	// minioClientSet is a clientset for our own API group
	minioClientSet clientset.Interface

	// statefulSetLister is able to list/get StatefulSets from a shared
	// informer's store.
	statefulSetLister appslisters.StatefulSetLister
	// statefulSetListerSynced returns true if the StatefulSet shared informer
	// has synced at least once.
	statefulSetListerSynced cache.InformerSynced

	// minioInstancesLister lists MinioInstance from a shared informer's
	// store.
	minioInstancesLister listers.MinioInstanceLister
	// minioInstancesSynced returns true if the StatefulSet shared informer
	// has synced at least once.
	minioInstancesSynced cache.InformerSynced

	// serviceLister is able to list/get Services from a shared informer's
	// store.
	serviceLister corelisters.ServiceLister
	// serviceListerSynced returns true if the Service shared informer
	// has synced at least once.
	serviceListerSynced cache.InformerSynced

	// queue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue queue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new sample controller
func NewController(
	kubeClientSet kubernetes.Interface,
	minioClientSet clientset.Interface,
	statefulSetInformer appsinformers.StatefulSetInformer,
	minioInstanceInformer informers.MinioInstanceInformer,
	serviceInformer coreinformers.ServiceInformer) *Controller {

	// Create event broadcaster
	// Add minio-controller types to the default Kubernetes Scheme so Events can be
	// logged for minio-controller types.
	minioscheme.AddToScheme(scheme.Scheme)
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClientSet.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeClientSet:           kubeClientSet,
		minioClientSet:          minioClientSet,
		statefulSetLister:       statefulSetInformer.Lister(),
		statefulSetListerSynced: statefulSetInformer.Informer().HasSynced,
		minioInstancesLister:    minioInstanceInformer.Lister(),
		minioInstancesSynced:    minioInstanceInformer.Informer().HasSynced,
		serviceLister:           serviceInformer.Lister(),
		serviceListerSynced:     serviceInformer.Informer().HasSynced,
		workqueue:               queue.NewNamedRateLimitingQueue(queue.DefaultControllerRateLimiter(), "MinioInstances"),
		recorder:                recorder,
	}

	glog.Info("Setting up event handlers")
	// Set up an event handler for when MinioInstance resources change
	minioInstanceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueMinioInstance,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueMinioInstance(new)
		},
	})
	// Set up an event handler for when StatefulSet resources change. This
	// handler will lookup the owner of the given StatefulSet, and if it is
	// owned by a MinioInstance resource will enqueue that MinioInstance resource for
	// processing. This way, we don't need to implement custom logic for
	// handling StatefulSet resources. More info on this pattern:
	// https://github.com/kubernetes/community/blob/8cafef897a22026d42f5e5bb3f104febe7e29830/contributors/devel/controllers.md
	statefulSetInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			newDepl := new.(*appsv1.StatefulSet)
			oldDepl := old.(*appsv1.StatefulSet)
			if newDepl.ResourceVersion == oldDepl.ResourceVersion {
				// Periodic resync will send update events for all known StatefulSet.
				// Two different versions of the same StatefulSet will always have different RVs.
				return
			}
			controller.handleObject(new)
		},
		DeleteFunc: controller.handleObject,
	})
	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it w	ill shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting MinioInstance controller")

	// Wait for the caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.statefulSetListerSynced, c.minioInstancesSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")
	// Launch two workers to process MinioInstance resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
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
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// MinioInstance resource to be synced.
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		glog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}
	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the MinioInstance resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("Invalid resource key: %s", key))
		return nil
	}

	nsName := types.NamespacedName{Namespace: namespace, Name: name}

	// Get the MinioInstance resource with this namespace/name
	mi, err := c.minioInstancesLister.MinioInstances(namespace).Get(name)
	if err != nil {
		// The MinioInstance resource may no longer exist, in which case we stop processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("MinioInstance '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}

	mi.EnsureDefaults()

	svc, err := c.serviceLister.Services(mi.Namespace).Get(mi.Name)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		glog.V(2).Infof("Creating a new Service for cluster %q", nsName)
		svc = services.NewForCluster(mi)
		_, err = c.kubeClientSet.CoreV1().Services(svc.Namespace).Create(svc)
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// Get the StatefulSet with the name specified in MinioInstance.spec
	ss, err := c.statefulSetLister.StatefulSets(mi.Namespace).Get(mi.Name)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		ss = statefulsets.NewForCluster(mi, svc.Name)
		_, err = c.kubeClientSet.AppsV1().StatefulSets(mi.Namespace).Create(ss)
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// If the StatefulSet is not controlled by this MinioInstance resource, we should log
	// a warning to the event recorder and ret
	if !metav1.IsControlledBy(ss, mi) {
		msg := fmt.Sprintf(MessageResourceExists, ss.Name)
		c.recorder.Event(mi, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}

	// If this number of the replicas on the MinioInstance resource is specified, and the
	// number does not equal the current desired replicas on the StatefulSet, we
	// should update the StatefulSet resource.
	if mi.Spec.Replicas != *ss.Spec.Replicas {
		glog.V(4).Infof("MinioInstance %s replicas: %d, StatefulSet replicas: %d", name, mi.Spec.Replicas, *ss.Spec.Replicas)
		ss = statefulsets.NewForCluster(mi, svc.Name)
		_, err = c.kubeClientSet.AppsV1().StatefulSets(mi.Namespace).Update(ss)
	}

	// If this container version on the MinioInstance resource is specified, and the
	// version does not equal the current desired version in the StatefulSet, we
	// should update the StatefulSet resource.
	currentVersion := strings.TrimPrefix(ss.Spec.Template.Spec.Containers[0].Image, constants.MinioImagePath+":")
	if mi.Spec.Version != currentVersion {
		glog.V(4).Infof("Updating MinioInstance %s Minio server version %d, to: %d", name, mi.Spec.Version, currentVersion)
		ss = statefulsets.NewForCluster(mi, svc.Name)
		_, err = c.kubeClientSet.AppsV1().StatefulSets(mi.Namespace).Update(ss)
	}

	// If an error occurs during Update, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// Finally, we update the status block of the MinioInstance resource to reflect the
	// current state of the world
	err = c.updateMinioInstanceStatus(mi, ss)
	if err != nil {
		return err
	}

	c.recorder.Event(mi, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) updateMinioInstanceStatus(minioInstance *miniov1beta1.MinioInstance, statefulSet *appsv1.StatefulSet) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	minioInstanceCopy := minioInstance.DeepCopy()
	minioInstanceCopy.Status.AvailableReplicas = statefulSet.Status.Replicas
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the MinioInstance resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := c.minioClientSet.MinioV1beta1().MinioInstances(minioInstance.Namespace).Update(minioInstanceCopy)
	return err
}

// enqueueMinioInstance takes a MinioInstance resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than MinioInstance.
func (c *Controller) enqueueMinioInstance(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the MinioInstance resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that MinioInstance resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		glog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	glog.V(4).Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a MinioInstance, we should not do anything more
		// with it.
		if ownerRef.Kind != "MinioInstance" {
			return
		}

		minioInstance, err := c.minioInstancesLister.MinioInstances(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			glog.V(4).Infof("ignoring orphaned object '%s' of minioInstance '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueueMinioInstance(minioInstance)
		return
	}
}
