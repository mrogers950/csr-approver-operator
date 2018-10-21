package csrapprover

import (
	"fmt"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	csrv1alpha1 "github.com/mrogers950/csr-approver-operator/pkg/apis/csrapprover.config.openshift.io/v1alpha1"
)

type CSRApproverOptions struct {
	Config *csrv1alpha1.CSRApproverConfig
}

func (o *CSRApproverOptions) RunCSRApprover(clientConfig *rest.Config, stopCh <-chan struct{}) error {
	kubeClient, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return err
	}
	kubeInformers := informers.NewSharedInformerFactory(kubeClient, 2*time.Minute)

	controllerOpts, err := NewControllerOptions(o.Config)
	if err != nil {
		return err
	}
	csrApproverController := NewCSRApproverController(
		controllerOpts,
		kubeClient.CertificatesV1beta1(),
		kubeInformers.Certificates().V1beta1().CertificateSigningRequests(),
		10*time.Minute,
	)

	kubeInformers.Start(stopCh)
	go csrApproverController.Run(1, stopCh)
	<-stopCh

	return fmt.Errorf("stopped")
}
