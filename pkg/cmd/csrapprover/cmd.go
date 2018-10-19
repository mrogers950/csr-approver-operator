package csrapprover

import (
	"math/rand"
	"os"
	"time"

	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/util/logs"

	"fmt"

	"github.com/golang/glog"
	csrv1alpha1 "github.com/mrogers950/csr-approver-operator/pkg/apis/csrapprover.config.openshift.io/v1alpha1"
	"github.com/mrogers950/csr-approver-operator/pkg/controller/csrapprover"
	"github.com/mrogers950/csr-approver-operator/pkg/version"
	operatorv1alpha1 "github.com/openshift/api/operator/v1alpha1"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/openshift/library-go/pkg/serviceability"
)

var (
	componentName = "openshift-csr-approver"
	configScheme  = runtime.NewScheme()
)

func init() {
	if err := operatorv1alpha1.AddToScheme(configScheme); err != nil {
		panic(err)
	}
	if err := csrv1alpha1.AddToScheme(configScheme); err != nil {
		panic(err)
	}
}

type ControllerCommandOptions struct {
	basicFlags *controllercmd.ControllerFlags
}

func NewController() *cobra.Command {
	o := &ControllerCommandOptions{
		basicFlags: controllercmd.NewControllerFlags(),
	}

	cmd := &cobra.Command{
		Use:   "csr-approver",
		Short: "Start the CSR API Approver",
		Run: func(cmd *cobra.Command, args []string) {
			// boiler plate for the "normal" command
			rand.Seed(time.Now().UTC().UnixNano())
			logs.InitLogs()
			defer logs.FlushLogs()
			defer serviceability.BehaviorOnPanic(os.Getenv("OPENSHIFT_ON_PANIC"), version.Get())()
			defer serviceability.Profile(os.Getenv("OPENSHIFT_PROFILE")).Stop()
			serviceability.StartProfiler()

			if err := o.basicFlags.Validate(); err != nil {
				glog.Fatal(err)
			}

			if err := o.StartController(); err != nil {
				glog.Fatal(err)
			}
		},
	}
	o.basicFlags.AddFlags(cmd)

	return cmd
}

// StartController runs the controller
func (o *ControllerCommandOptions) StartController() error {
	uncastConfig, err := o.basicFlags.ToConfigObj(configScheme, csrv1alpha1.SchemeGroupVersion)
	if err != nil {
		return err
	}

	config, ok := uncastConfig.(*csrv1alpha1.CSRApproverConfig)
	if !ok {
		return fmt.Errorf("unexpected config: %T", uncastConfig)
	}

	return controllercmd.NewController(componentName, (&csrapprover.CSRApproverOptions{Config: config}).RunCSRApprover).
		WithKubeConfigFile(o.basicFlags.KubeConfigFile, nil).
		//WithLeaderElection(config.LeaderElection, "", componentName+"-lock").
		Run()
}
