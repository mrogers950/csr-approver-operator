package operator

import (
	"fmt"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/mrogers950/csr-approver-operator/pkg/client/clientset/versioned"
	csrv1alpha1informer "github.com/mrogers950/csr-approver-operator/pkg/client/informers/externalversions"
)

func RunOperator(clientConfig *rest.Config, stopCh <-chan struct{}) error {
	kubeClient, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		panic(err)
	}
	csrClient, err := versioned.NewForConfig(clientConfig)
	if err != nil {
		panic(err)
	}

	operatorInformers := csrv1alpha1informer.NewSharedInformerFactory(csrClient, 10*time.Minute)
	kubeInformersNamespaced := informers.NewFilteredSharedInformerFactory(kubeClient, 10*time.Minute, targetNamespaceName, nil)

	operator := NewCSRApproverOperator(
		kubeInformersNamespaced,
		csrClient.CsrapproverV1alpha1(),
		operatorInformers.Csrapprover().V1alpha1().CSRApproverOperatorConfigs(),
		kubeClient.AppsV1(),
		kubeClient.CoreV1(),
		kubeClient.RbacV1(),
	)

	operatorInformers.Start(stopCh)
	kubeInformersNamespaced.Start(stopCh)

	operator.Run(1, stopCh)
	return fmt.Errorf("stopped")
}
