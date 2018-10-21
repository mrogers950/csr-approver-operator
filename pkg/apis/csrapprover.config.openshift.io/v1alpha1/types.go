package v1alpha1

import (
	operatorsv1alpha1api "github.com/openshift/api/operator/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CSRApproverConfig is the configuration for the CSR Approval Controller
type CSRApproverConfig struct {
	metav1.TypeMeta `json:",inline"`
	// TODO: more parameters
	Profiles []CSRApprovalProfile `json:"profiles"`
}

type CSRApprovalProfile struct {
	Name            string   `json:"name"`
	AllowedNames    []string `json:"allowedNames"`
	AllowedSubjects []string `json:"allowedSubjects"`
	AllowedUsages   []string `json:"allowedUsages"`
	AllowedUsers    []string `json:"allowedUsers"`
	AllowedGroups   []string `json:"allowedGroups"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CSRApproverOperatorConfig describes configuration for CSRApproverOperator
type CSRApproverOperatorConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CSRApproverOperatorConfigSpec   `json:"spec"`
	Status CSRApproverOperatorConfigStatus `json:"status"`
}

// CSRApproverOperatorConfigSpec is the spec field for the CSR Approver configuration
type CSRApproverOperatorConfigSpec struct {
	operatorsv1alpha1api.OperatorSpec `json:",inline"`

	CSRApproverConfig runtime.RawExtension `json:"csrApproverConfig"`
}

// CSRApproverOperatorConfigStatus is the status of the CSR Approver operator
type CSRApproverOperatorConfigStatus struct {
	// TODO: what fields do we need?
	operatorsv1alpha1api.OperatorStatus `json:",inline"`

	//ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CSRApproverOperatorConfigList is a list of CSR Approver Operator configurations
type CSRApproverOperatorConfigList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items contains the items
	Items []CSRApproverOperatorConfig `json:"items"`
}
