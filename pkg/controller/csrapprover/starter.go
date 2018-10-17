package csrapprover

import (
	"fmt"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"k8s.io/client-go/informers"
)

type CSRApproverConfig struct {
}

type CSRApproverOptions struct {
	Config CSRApproverConfig
}

func (o *CSRApproverOptions) RunCSRApprover(clientConfig *rest.Config, stopCh <-chan struct{}) error {
	kubeClient, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return err
	}
	kubeInformers := informers.NewSharedInformerFactory(kubeClient, 2*time.Minute)

	csrApproverController := NewCSRApproverController()

	kubeInformers.Start(stopCh)
	go csrApproverController.Run(1, stopCh)
	<-stopCh

	return fmt.Errorf("stopped")
}
