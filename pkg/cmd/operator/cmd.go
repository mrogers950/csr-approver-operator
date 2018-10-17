package operator

import (
	"github.com/spf13/cobra"

	"github.com/mrogers950/csr-approver-operator/pkg/operator"
	opversion "github.com/mrogers950/csr-approver-operator/pkg/version"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
)

func NewOperator() *cobra.Command {
	cmd := controllercmd.
		NewControllerCommandConfig("openshift-csr-approver-operator", opversion.Get(), operator.RunOperator).
		NewCommand()
	cmd.Use = "operator"
	cmd.Short = "Start the CSR API Approver Operator"

	return cmd
}
