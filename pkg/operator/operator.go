package operator

import (
	"fmt"
	"time"

	//"github.com/blang/semver"
	"github.com/golang/glog"

	//corev1 "k8s.io/api/core/v1"
	//apierrors "k8s.io/apimachinery/pkg/api/errors"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	//"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	//"k8s.io/client-go/informers"
	//appsclientv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	//coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	//rbacclientv1 "k8s.io/client-go/kubernetes/typed/rbac/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	//operatorsv1alpha1 "github.com/openshift/api/operator/v1alpha1"
	//scsclientv1alpha1 "github.com/openshift/client-go/servicecertsigner/clientset/versioned/typed/servicecertsigner/v1alpha1"
	//scsinformerv1alpha1 "github.com/openshift/client-go/servicecertsigner/informers/externalversions/servicecertsigner/v1alpha1"
	//"github.com/openshift/library-go/pkg/operator/v1alpha1helpers"
	//"github.com/openshift/library-go/pkg/operator/versioning"
	"k8s.io/client-go/rest"
)

const (
	targetNamespaceName = "openshift-service-cert-signer"
	workQueueKey        = "key"
)

type CSRApproverOperator struct {
	operatorConfigClient rest.RESTClient
	// + operatorConfigClient CSRApproverOperatorConfigGetter

	// + csrClient csrclientinterface

	// queue only ever has one item, but it has nice error handling backoff/retry semantics
	queue workqueue.RateLimitingInterface
}

func NewCSRApproverOperator(operatorConfigClient rest.RESTClient) *CSRApproverOperator {
	c := &CSRApproverOperator{
		operatorConfigClient: operatorConfigClient,
		// csrClient:
		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "CSRApproverOperator"),
	}

	// Here we add event handlers to our informers.
	//serviceCertSignerConfigInformer.Informer().AddEventHandler(c.eventHandler())
	//namespacedKubeInformers.Core().V1().ConfigMaps().Informer().AddEventHandler(c.eventHandler())
	//namespacedKubeInformers.Core().V1().ServiceAccounts().Informer().AddEventHandler(c.eventHandler())
	//namespacedKubeInformers.Core().V1().Services().Informer().AddEventHandler(c.eventHandler())
	//namespacedKubeInformers.Apps().V1().Deployments().Informer().AddEventHandler(c.eventHandler())

	return c
}

func (c CSRApproverOperator) syncCSRApproverOperatorConfig() error {
	// get operator config instance

	// check management state. Removed, Unmanaged

	// check version

	// sync_v311_00_to_latest
	// update cm, sa, services, dep

	// update status
	return nil

}

// Run starts the serviceCertSigner and blocks until stopCh is closed.
func (c *CSRApproverOperator) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	glog.Infof("Starting CSRApproverOperator")
	defer glog.Infof("Shutting down CSRApproverOperator")

	// doesn't matter what workers say, only start one.
	go wait.Until(c.runWorker, time.Second, stopCh)

	<-stopCh
}

func (c *CSRApproverOperator) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *CSRApproverOperator) processNextWorkItem() bool {
	dsKey, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(dsKey)

	err := c.syncCSRApproverOperatorConfig()
	if err == nil {
		c.queue.Forget(dsKey)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("%v failed with : %v", dsKey, err))
	c.queue.AddRateLimited(dsKey)

	return true
}

// eventHandler queues the operator to check spec and status
func (c *CSRApproverOperator) eventHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { c.queue.Add(workQueueKey) },
		UpdateFunc: func(old, new interface{}) { c.queue.Add(workQueueKey) },
		DeleteFunc: func(obj interface{}) { c.queue.Add(workQueueKey) },
	}
}
