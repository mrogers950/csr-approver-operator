package csrapprover

import (
	"fmt"
	"time"

	"github.com/golang/glog"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
)

//const (
//	// ServingCertSecretAnnotation stores the name of the secret to generate into.
//	//ServingCertSecretAnnotation = "service.alpha.openshift.io/serving-cert-secret-name"
//	// ServingCertCreatedByAnnotation stores the of the signer common name.  This could be used later to see if the
//	// services need to have the the serving certs regenerated.  The presence and matching of this annotation prevents
//	// regeneration
//	ServingCertCreatedByAnnotation = "service.alpha.openshift.io/serving-cert-signed-by"
//	// ServingCertErrorAnnotation stores the error that caused cert generation failures.
//	ServingCertErrorAnnotation = "service.alpha.openshift.io/serving-cert-generation-error"
//	// ServingCertErrorNumAnnotation stores how many consecutive errors we've hit.  A value of the maxRetries will prevent
//	// the controller from reattempting until it is cleared.
//	ServingCertErrorNumAnnotation = "service.alpha.openshift.io/serving-cert-generation-error-num"
//	// ServiceUIDAnnotation is an annotation on a secret that indicates which service created it, by UID
//	ServiceUIDAnnotation = "service.alpha.openshift.io/originating-service-uid"
//	// ServiceNameAnnotation is an annotation on a secret that indicates which service created it, by Name to allow reverse lookups on services
//	// for comparison against UIDs
//	ServiceNameAnnotation = "service.alpha.openshift.io/originating-service-name"
//	// ServingCertExpiryAnnotation is an annotation that holds the expiry time of the certificate.  It accepts time in the
//	// RFC3339 format: 2018-11-29T17:44:39Z
//	ServingCertExpiryAnnotation = "service.alpha.openshift.io/expiry"
//)

// CSRApproverController is responsible for approval of CSR requests based on the configured attrubute ACL
type CSRApproverController struct {
	// client csrclient
	// CSRs that need to be checked
	queue      workqueue.RateLimitingInterface
	maxRetries int
	//csrLister    listers.ServiceLister
	//csrHasSynced cache.InformerSynced
	// syncHandler does the work. It's factored out for unit testing
	syncHandler func(csrKey string) error
}

// NewCSRApproverController creates a new CSRApproverController.
// TODO this should accept a shared informer
func NewCSRApproverController() *CSRApproverController {
	sc := &CSRApproverController{
		queue:      workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		maxRetries: 10,
	}

	//services.Informer().AddEventHandlerWithResyncPeriod(
	//	cache.ResourceEventHandlerFuncs{
	//		AddFunc: func(obj interface{}) {
	//			service := obj.(*v1.Service)
	//			glog.V(4).Infof("Adding service %s", service.Name)
	//			sc.enqueueService(obj)
	//		},
	//		UpdateFunc: func(old, cur interface{}) {
	//			service := cur.(*v1.Service)
	//			glog.V(4).Infof("Updating service %s", service.Name)
	//			// Resync on service object relist.
	//			sc.enqueueService(cur)
	//		},
	//	},
	//	resyncInterval,
	//)
	//sc.serviceLister = services.Lister()
	//sc.serviceHasSynced = services.Informer().GetController().HasSynced
	//
	//secrets.Informer().AddEventHandlerWithResyncPeriod(
	//	cache.ResourceEventHandlerFuncs{
	//		DeleteFunc: sc.deleteSecret,
	//	},
	//	resyncInterval,
	//)
	//sc.secretHasSynced = services.Informer().GetController().HasSynced
	//sc.secretLister = secrets.Lister()

	sc.syncHandler = sc.syncCSR

	return sc
}

// Run begins watching and syncing.
func (sc *CSRApproverController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer sc.queue.ShutDown()

	//if !cache.WaitForCacheSync(stopCh, sc.serviceHasSynced, sc.secretHasSynced) {
	//	return
	//}

	glog.V(5).Infof("Starting workers")
	for i := 0; i < workers; i++ {
		go wait.Until(sc.worker, time.Second, stopCh)
	}
	<-stopCh
	glog.V(1).Infof("Shutting down")
}

// deleteSecret handles the case when the service certificate secret is manually removed.
// In that case the secret will be automatically recreated.
func (sc *CSRApproverController) deleteCSR(obj interface{}) {
	//secret, ok := obj.(*v1.Secret)
	//if !ok {
	//	return
	//}
	//if _, exists := secret.Annotations[ServiceNameAnnotation]; !exists {
	//	return
	//}
	//service, err := sc.serviceLister.Services(secret.Namespace).Get(secret.Annotations[ServiceNameAnnotation])
	//if kapierrors.IsNotFound(err) {
	//	return
	//}
	//if err != nil {
	//	utilruntime.HandleError(fmt.Errorf("Unable to get service %s/%s: %v", secret.Namespace, secret.Annotations[ServiceNameAnnotation], err))
	//	return
	//}
	//glog.V(4).Infof("Recreating secret for service %q", service.Namespace+"/"+service.Name)
	//sc.enqueueService(service)
}

func (sc *CSRApproverController) enqueueCSR(obj interface{}) {
	//_, ok := obj.(*v1.Service)
	//if !ok {
	//	return
	//}
	//key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	//if err != nil {
	//	glog.Errorf("Couldn't get key for object %+v: %v", obj, err)
	//	return
	//}
	//
	//sc.queue.Add(key)
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (sc *CSRApproverController) worker() {
	for sc.work() {
	}
}

// work returns true if the worker thread should continue
func (sc *CSRApproverController) work() bool {
	key, quit := sc.queue.Get()
	if quit {
		return false
	}
	defer sc.queue.Done(key)

	if err := sc.syncHandler(key.(string)); err == nil {
		// this means the request was successfully handled.  We should "forget" the item so that any retry
		// later on is reset
		sc.queue.Forget(key)

	} else {
		// if we had an error it means that we didn't handle it, which means that we want to requeue the work
		utilruntime.HandleError(fmt.Errorf("error syncing service, it will be retried: %v", err))
		sc.queue.AddRateLimited(key)
	}

	return true
}

func (sc *CSRApproverController) syncCSR(key string) error {
	return nil
}
