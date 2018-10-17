package csrapprover

import (
	"math/rand"
	"os"
	"time"

	"github.com/spf13/cobra"

	"k8s.io/apiserver/pkg/util/logs"

	"github.com/golang/glog"
	//operatorv1alpha1 "github.com/openshift/api/operator/v1alpha1"
	//servicecertsignerv1alpha1 "github.com/openshift/api/servicecertsigner/v1alpha1"
	"github.com/mrogers950/csr-approver-operator/pkg/controller/csrapprover"
	"github.com/mrogers950/csr-approver-operator/pkg/version"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/openshift/library-go/pkg/serviceability"
)

var (
	componentName = "openshift-csr-approver"
)

func init() {
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

	// create config to go in approver options

	config := csrapprover.CSRApproverConfig{}
	return controllercmd.NewController(componentName, (&csrapprover.CSRApproverOptions{Config: config}).RunCSRApprover).
		WithKubeConfigFile(o.basicFlags.KubeConfigFile, nil).
		//WithLeaderElection(config.LeaderElection, "", componentName+"-lock").
		Run()
}
