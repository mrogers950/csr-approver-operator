package operator

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	appsclientv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"

	csrv1alpha1api "github.com/mrogers950/csr-approver-operator/pkg/apis/csrapprover.config.openshift.io/v1alpha1"
	"github.com/mrogers950/csr-approver-operator/pkg/operator/v311_00_assets"
	operatorsv1alpha1 "github.com/openshift/api/operator/v1alpha1"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	"github.com/openshift/library-go/pkg/operator/resource/resourcemerge"
	"github.com/openshift/library-go/pkg/operator/resource/resourceread"
)

// syncCSRApproverOperator synchronizes the CSR Approver Operator, but does not manage an upgrade.
func syncCSRApproverOperator(c CSRApproverOperator, operatorConfig *csrv1alpha1api.CSRApproverOperatorConfig, previousAvailability *operatorsv1alpha1.VersionAvailablity) (operatorsv1alpha1.VersionAvailablity, []error) {
	csrApproverAvailability, csrApproverErrors := syncCSRApproverController(c, operatorConfig, previousAvailability)

	allErrors := []error{}
	allErrors = append(allErrors, csrApproverErrors...)

	mergedVersionAvailability := operatorsv1alpha1.VersionAvailablity{
		Version: operatorConfig.Spec.Version,
	}
	mergedVersionAvailability.Generations = append(mergedVersionAvailability.Generations, csrApproverAvailability.Generations...)
	if csrApproverAvailability.UpdatedReplicas > 0 {
		mergedVersionAvailability.UpdatedReplicas = 1
	}
	if csrApproverAvailability.ReadyReplicas > 0 {
		mergedVersionAvailability.ReadyReplicas = 1
	}
	for _, err := range allErrors {
		mergedVersionAvailability.Errors = append(mergedVersionAvailability.Errors, err.Error())
	}

	return mergedVersionAvailability, allErrors
}

// syncSigningController_v311_00_to_latest takes care of synchronizing (not upgrading) the thing we're managing.
// most of the time the sync method will be good for a large span of minor versions
func syncCSRApproverController(c CSRApproverOperator, operatorConfig *csrv1alpha1api.CSRApproverOperatorConfig, previousAvailability *operatorsv1alpha1.VersionAvailablity) (operatorsv1alpha1.VersionAvailablity, []error) {
	versionAvailability := operatorsv1alpha1.VersionAvailablity{
		Version: operatorConfig.Spec.Version,
	}

	errors := []error{}
	var err error

	requiredNamespace := resourceread.ReadNamespaceV1OrDie(v311_00_assets.MustAsset("v3.11.0/csr-approver-controller/ns.yaml"))
	_, _, err = resourceapply.ApplyNamespace(c.corev1Client, requiredNamespace)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q: %v", "ns", err))
	}

	requiredClusterRole := resourceread.ReadClusterRoleV1OrDie(v311_00_assets.MustAsset("v3.11.0/csr-approver-controller/clusterrole.yaml"))
	_, _, err = resourceapply.ApplyClusterRole(c.rbacv1Client, requiredClusterRole)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q: %v", "clusterrole", err))
	}

	requiredClusterRoleBinding := resourceread.ReadClusterRoleBindingV1OrDie(v311_00_assets.MustAsset("v3.11.0/csr-approver-controller/clusterrolebinding.yaml"))
	_, _, err = resourceapply.ApplyClusterRoleBinding(c.rbacv1Client, requiredClusterRoleBinding)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q: %v", "clusterrolebinding", err))
	}

	requiredService := resourceread.ReadServiceV1OrDie(v311_00_assets.MustAsset("v3.11.0/csr-approver-controller/svc.yaml"))
	_, _, err = resourceapply.ApplyService(c.corev1Client, requiredService)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q: %v", "svc", err))
	}

	requiredSA := resourceread.ReadServiceAccountV1OrDie(v311_00_assets.MustAsset("v3.11.0/csr-approver-controller/sa.yaml"))
	_, saModified, err := resourceapply.ApplyServiceAccount(c.corev1Client, requiredSA)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q: %v", "sa", err))
	}

	_, configMapModified, err := manageControllerConfigMap(c.corev1Client, operatorConfig)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q: %v", "configmap", err))
	}

	forceDeployment := operatorConfig.ObjectMeta.Generation != operatorConfig.Status.ObservedGeneration
	if saModified { // SA modification can cause new tokens
		forceDeployment = true
	}
	if configMapModified {
		forceDeployment = true
	}

	actualDeployment, _, err := manageControllerDeployment(c.appsv1Client, operatorConfig, previousAvailability, forceDeployment)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q: %v", "deployment", err))
	}

	return resourcemerge.ApplyGenerationAvailability(versionAvailability, actualDeployment, errors...), errors
}

func manageControllerConfigMap(client coreclientv1.ConfigMapsGetter, operatorConfig *csrv1alpha1api.CSRApproverOperatorConfig) (*corev1.ConfigMap, bool, error) {
	configMap := resourceread.ReadConfigMapV1OrDie(v311_00_assets.MustAsset("v3.11.0/csr-approver-controller/cm.yaml"))
	defaultConfig := v311_00_assets.MustAsset("v3.11.0/csr-approver-controller/defaultconfig.yaml")
	requiredConfigMap, _, err := resourcemerge.MergeConfigMap(configMap, "controller-config.yaml", nil, defaultConfig, operatorConfig.Spec.CSRApproverConfig.Raw)
	if err != nil {
		return nil, false, err
	}
	return resourceapply.ApplyConfigMap(client, requiredConfigMap)
}

func manageControllerDeployment(client appsclientv1.DeploymentsGetter, operatorConfig *csrv1alpha1api.CSRApproverOperatorConfig, previousAvailability *operatorsv1alpha1.VersionAvailablity, forceDeployment bool) (*appsv1.Deployment, bool, error) {
	required := resourceread.ReadDeploymentV1OrDie(v311_00_assets.MustAsset("v3.11.0/csr-approver-controller/deployment.yaml"))
	required.Spec.Template.Spec.Containers[0].Image = operatorConfig.Spec.ImagePullSpec
	required.Spec.Template.Spec.Containers[0].Args = append(required.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("-v=%d", operatorConfig.Spec.Logging.Level))

	return resourceapply.ApplyDeployment(client, required, resourcemerge.ExpectedDeploymentGeneration(required, previousAvailability), forceDeployment)
}
