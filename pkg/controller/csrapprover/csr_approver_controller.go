package csrapprover

import (
	"fmt"
	"time"

	"github.com/golang/glog"

	"github.com/mrogers950/csr-approver-operator/pkg/apis/csrapprover.config.openshift.io/v1alpha1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"

	certv1beta1 "k8s.io/client-go/kubernetes/typed/certificates/v1beta1"

	v1beta12 "k8s.io/api/certificates/v1beta1"
	kapierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers/certificates/v1beta1"
	listers "k8s.io/client-go/listers/certificates/v1beta1"
	"k8s.io/client-go/tools/cache"
)

// CSRApproverController is responsible for approval of CSR requests based on the configured attrubute ACL
type CSRApproverController struct {
	csrClient certv1beta1.CertificateSigningRequestsGetter
	// CSRs that need to be checked
	queue      workqueue.RateLimitingInterface
	maxRetries int

	csrLister    listers.CertificateSigningRequestLister
	csrHasSynced cache.InformerSynced

	// syncHandler does the work. It's factored out for unit testing
	syncHandler func(csrKey string) error

	config *v1alpha1.CSRApproverConfig
}

// NewCSRApproverController creates a new CSRApproverController.
func NewCSRApproverController(
	controllerConfig *v1alpha1.CSRApproverConfig,
	csrClient certv1beta1.CertificateSigningRequestsGetter,
	csrInformer v1beta1.CertificateSigningRequestInformer,
	resyncInterval time.Duration,
) *CSRApproverController {
	sc := &CSRApproverController{
		queue:      workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		maxRetries: 10,
		config:     controllerConfig,
		csrClient:  csrClient,
	}

	csrInformer.Informer().AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				csr := obj.(*v1beta12.CertificateSigningRequest)
				glog.V(4).Infof("Adding CSR %s", csr.Name)
				sc.enqueueCSR(obj)
			},
			UpdateFunc: func(old, cur interface{}) {
				csr := cur.(*v1beta12.CertificateSigningRequest)
				glog.V(4).Infof("Updating CSR %s", csr.Name)
				sc.enqueueCSR(cur)
			},
		},
		resyncInterval,
	)
	sc.csrLister = csrInformer.Lister()
	sc.csrHasSynced = csrInformer.Informer().GetController().HasSynced

	sc.syncHandler = sc.syncCSR

	return sc
}

// Run begins watching and syncing.
func (sc *CSRApproverController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer sc.queue.ShutDown()

	if !cache.WaitForCacheSync(stopCh, sc.csrHasSynced) {
		return
	}

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
	_, ok := obj.(*v1beta12.CertificateSigningRequest)
	if !ok {
		return
	}
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		glog.Errorf("Couldn't get key for object %+v: %v", obj, err)
		return
	}

	sc.queue.Add(key)
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
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	cacheCsr, err := sc.csrLister.Get(name)
	if kapierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	csr := cacheCsr.DeepCopy()

	_ := &v1beta12.CertificateSigningRequest{
		TypeMeta: v1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:            "",
			GenerateName:    "",
			Namespace:       "",
			SelfLink:        "",
			UID:             "",
			ResourceVersion: "",
			Generation:      0,
			CreationTimestamp: v1.Time{
				Time: time.Time{},
			},
			DeletionTimestamp: &v1.Time{
				Time: time.Time{},
			},
			DeletionGracePeriodSeconds: nil,
			Labels:          nil,
			Annotations:     nil,
			OwnerReferences: nil,
			Initializers: &v1.Initializers{
				Pending: nil,
				Result: &v1.Status{
					TypeMeta: v1.TypeMeta{
						Kind:       "",
						APIVersion: "",
					},
					ListMeta: v1.ListMeta{
						SelfLink:        "",
						ResourceVersion: "",
						Continue:        "",
					},
					Status:  "",
					Message: "",
					Reason:  "",
					Details: &v1.StatusDetails{
						Name:              "",
						Group:             "",
						Kind:              "",
						UID:               "",
						Causes:            nil,
						RetryAfterSeconds: 0,
					},
					Code: 0,
				},
			},
			Finalizers:  nil,
			ClusterName: "",
		},
		Spec: v1beta12.CertificateSigningRequestSpec{
			Request:  nil,
			Usages:   nil,
			Username: "",
			UID:      "",
			Groups:   nil,
			Extra:    nil,
		},
		Status: v1beta12.CertificateSigningRequestStatus{
			Conditions:  nil,
			Certificate: nil,
		},
	}

	glog.Infof("sync CSR - CSR: %s, spec: %#v", csr.Name, csr.Spec)

	for _, cond := range csr.Status.Conditions {
		if cond.Type == v1beta12.CertificateApproved || cond.Type == v1beta12.CertificateDenied {
			glog.Infof("sync CSR - certificate is approved/denied, returning nil")
			return nil
		}
	}

	for _, allowed := range sc.config.AllowedUsages {
		seenUsage := false
		for _, usage := range csr.Spec.Usages {
			if allowed == string(usage) {
				seenUsage = true
				break
			}
		}
		if !seenUsage {
			return fmt.Errorf("boom usage not allowed")
		}
	}

	names, err := collectNamesFromCSR(csr.Spec)
	if err != nil {
		return err
	}

	groups, err := getCSRGroups(csr.Spec)
	if err != nil {
		return err
	}

	err = sc.Allowed(names, groups)
	if err != nil {
		return err
	}

	csr.Status.Conditions = []v1beta12.CertificateSigningRequestCondition{
		{
			Type:           v1beta12.CertificateApproved,
			Reason:         "Usage checks succeeded",
			Message:        "Approved by the OpenShift CSR Approver",
			LastUpdateTime: v1.Time{Time: time.Now()},
		},
	}
	_, err = sc.csrClient.CertificateSigningRequests().Update(csr)
	if err != nil {
		return err
	}

	return nil
}

func convertStringstoKeyUsages(usageStrings []string) ([]v1beta12.KeyUsage, error) {
	for _, s := range usageStrings {

	}
}
