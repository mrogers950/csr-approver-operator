package csrapprover

import (
	"crypto/x509"
	"fmt"
	"time"

	"github.com/golang/glog"

	certapi "k8s.io/api/certificates/v1beta1"
	kapierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers/certificates/v1beta1"
	certv1beta1 "k8s.io/client-go/kubernetes/typed/certificates/v1beta1"
	listers "k8s.io/client-go/listers/certificates/v1beta1"
	"k8s.io/client-go/tools/cache"
	csrutil "k8s.io/client-go/util/certificate/csr"
	"k8s.io/client-go/util/workqueue"

	"github.com/mrogers950/csr-approver-operator/pkg/apis/csrapprover.config.openshift.io/v1alpha1"
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

	config *controllerConfig
}

type controllerConfig struct {
	profiles map[string]permissionProfile
}

// How permission profiles should work:
// The controller has 0 permission profiles. Auto-deny everything.
// The controller has one permission profile named "INSECURE-AUTO-APPROVE". Auto-approve everything.
// The controller has one or more named permission profiles. A csr must satsify at least one of the profiles.
type permissionProfile struct {
	allowedNames    []string
	allowedUsages   []certapi.KeyUsage
	allowedUsers    []string
	allowedGroups   []string
	allowedSubjects []string
}

func isValidUsage(usage string) bool {
	switch certapi.KeyUsage(usage) {
	case certapi.UsageClientAuth:
	case certapi.UsageServerAuth:
	case certapi.UsageKeyEncipherment:
	case certapi.UsageDataEncipherment:
	case certapi.UsageDigitalSignature:
	case certapi.UsageKeyAgreement:
	case certapi.UsageCertSign:
	case certapi.UsageSigning:
	case certapi.UsageEncipherOnly:
	case certapi.UsageAny:
	case certapi.UsageCodeSigning:
	case certapi.UsageContentCommittment:
	case certapi.UsageCRLSign:
	case certapi.UsageDecipherOnly:
	case certapi.UsageEmailProtection:
	case certapi.UsageIPsecEndSystem:
	case certapi.UsageIPsecTunnel:
	case certapi.UsageIPsecUser:
	case certapi.UsageMicrosoftSGC:
	case certapi.UsageNetscapSGC:
	case certapi.UsageOCSPSigning:
	case certapi.UsageSMIME:
	case certapi.UsageTimestamping:
	default:
		return false
	}
	return true
}

func NewControllerOptions(config *v1alpha1.CSRApproverConfig) (*controllerConfig, error) {
	profiles := make(map[string]permissionProfile, 0)

	// Validate/convert the controller config.
	for _, profile := range config.Profiles {
		p := permissionProfile{}
		if profile.Name == "" {
			continue
		}
		if _, ok := profiles[profile.Name]; ok {
			return nil, fmt.Errorf("Duplicate allow profiles configured: \"%s\"", profile.Name)
		}
		p.allowedUsages = make([]certapi.KeyUsage, 0)
		for _, usage := range profile.AllowedUsages {
			if !isValidUsage(usage) {
				return nil, fmt.Errorf("Not a supported certificate Usage: \"%s\"", usage)
			}
			p.allowedUsages = append(p.allowedUsages, certapi.KeyUsage(usage))
		}
		p.allowedNames = profile.AllowedNames
		p.allowedGroups = profile.AllowedGroups
		p.allowedSubjects = profile.AllowedSubjects
		p.allowedUsers = profile.AllowedUsers
		profiles[profile.Name] = p
	}

	return &controllerConfig{
		profiles: profiles,
	}, nil
}

// NewCSRApproverController creates a new CSRApproverController.
func NewCSRApproverController(
	controllerConfig *controllerConfig,
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
				csr := obj.(*certapi.CertificateSigningRequest)
				glog.V(4).Infof("Queueing CSR add: %s", csr.Name)
				sc.enqueueCSR(obj)
			},
			UpdateFunc: func(old, cur interface{}) {
				csr := cur.(*certapi.CertificateSigningRequest)
				glog.V(4).Infof("Queueing CSR update: %s", csr.Name)
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

func (sc *CSRApproverController) enqueueCSR(obj interface{}) {
	_, ok := obj.(*certapi.CertificateSigningRequest)
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

	glog.V(4).Infof("Checking CSR %s, Spec: %#v", csr.Name, csr.Spec)

	for _, cond := range csr.Status.Conditions {
		if cond.Type == certapi.CertificateApproved || cond.Type == certapi.CertificateDenied {
			glog.V(4).Infof("CSR %s already handled", csr.Name)
			return nil
		}
	}

	parsedCsr, err := csrutil.ParseCSR(csr)
	if err != nil {
		glog.V(4).Infof("CSR %s failed to parse: %v", err)
		return sc.denyCSR(csr, "Badly formed CSR")
	}

	if !allowedByProfiles(sc.config.profiles, csr.Spec, parsedCsr) {
		// TODO: give a failure hint
		glog.V(4).Infof("CSR %s denied", csr.Name)
		return sc.denyCSR(csr, "Not allowed by any approval profile")
	}

	glog.V(4).Infof("CSR %s approved", csr.Name)
	return sc.approveCSR(csr, "CSR Authorized")
}

const InsecureProfileName = "INSECURE-AUTO-APPROVE"

func allowedByProfiles(profiles map[string]permissionProfile, spec certapi.CertificateSigningRequestSpec, csr *x509.CertificateRequest) bool {
	if len(profiles) == 0 {
		// Auto-deny
		return false
	}

	for name, profile := range profiles {
		// A single insecure auto-approver profile.
		if name == InsecureProfileName && len(profiles) == 1 {
			return true
		}

		if !profile.csrUsageAllowed(spec.Usages) {
			continue
		}
		if !profile.csrGroupsAllowed(spec.Groups) {
			continue
		}
		if !profile.csrUserAllowed(spec.Username) {
			continue
		}
		if !profile.csrSubjectAllowed(csr) {
			continue
		}
		if !profile.csrNamesAllowed(csr.DNSNames) {
			continue
		}
		// All checks succeeded, allowed under this profile.
		return true
	}
	// Not covered under any profile.
	return false
}

func (p *permissionProfile) csrSubjectAllowed(csr *x509.CertificateRequest) bool {
	if len(p.allowedSubjects) == 0 {
		// No restriction
		return true
	}
	// TODO: Match wildcard DN
	for i := range p.allowedSubjects {
		if csr.Subject.String() == p.allowedSubjects[i] {
			return true
		}
	}
	return false
}

func (p *permissionProfile) csrUserAllowed(user string) bool {
	if len(p.allowedUsers) == 0 {
		// No restriction
		return true
	}
	for i := range p.allowedUsers {
		if user == p.allowedUsers[i] {
			return true
		}
	}
	return false
}

func (p *permissionProfile) csrGroupsAllowed(groups []string) bool {
	if len(p.allowedGroups) == 0 {
		// No restriction
		return true
	}
	if !subset(groups, p.allowedGroups) {
		return false
	}
	return true
}
func (p *permissionProfile) csrNamesAllowed(names []string) bool {
	if len(p.allowedNames) == 0 {
		// No restriction
		return true
	}
	// TODO: Match wildcard names and IP address names
	if !subset(names, p.allowedNames) {
		return false
	}
	return true
}

func (p *permissionProfile) csrUsageAllowed(usages []certapi.KeyUsage) bool {
	if len(p.allowedUsages) == 0 {
		// No restriction.
		return true
	}

	if !usageSubset(usages, p.allowedUsages) {
		return false
	}
	return true
}

func (sc *CSRApproverController) approveCSR(csr *certapi.CertificateSigningRequest, approvalReason string) error {
	csr.Status.Conditions = []certapi.CertificateSigningRequestCondition{
		{
			Type:           certapi.CertificateApproved,
			Reason:         approvalReason,
			Message:        "Approved by the OpenShift CSR Approver",
			LastUpdateTime: v1.Time{Time: time.Now()},
		},
	}
	_, err := sc.csrClient.CertificateSigningRequests().Update(csr)
	return err
}

func (sc *CSRApproverController) denyCSR(csr *certapi.CertificateSigningRequest, denyReason string) error {
	csr.Status.Conditions = []certapi.CertificateSigningRequestCondition{
		{
			Type:           certapi.CertificateDenied,
			Reason:         denyReason,
			Message:        "Denied by the OpenShift CSR Approver",
			LastUpdateTime: v1.Time{Time: time.Now()},
		},
	}
	_, err := sc.csrClient.CertificateSigningRequests().Update(csr)
	return err
}
