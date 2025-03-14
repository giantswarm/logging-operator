package vintagemc

import (
	appv1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

type Object struct {
	client.Object
	*loggedcluster.LoggingAgent
	Options loggedcluster.Options
}

func (o Object) HasLoggingEnabled() bool {
	return o.Options.EnableLoggingFlag
}

func (o Object) IsInsecureCA() bool {
	return o.Options.InsecureCA
}

func (o Object) GetAppsNamespace() string {
	return "giantswarm"
}

func (o Object) AppConfigName(app string) string {
	return app
}

func (o Object) GetClusterName() string {
	// return installation name for vintage MC
	return o.Options.InstallationName
}

func (o Object) GetInstallationName() string {
	return o.Options.InstallationName
}

func (o Object) GetEnableLoggingFlag() bool {
	return o.Options.EnableLoggingFlag
}

func (o Object) GetObject() client.Object {
	return o.Object
}

func (o Object) GetTenant() string {
	return o.GetClusterName()
}

func (o Object) IsCAPI() bool {
	return false
}

// On vintage MC, there's no support for extraconfig so we should use standard user values
func (o Object) GetObservabilityBundleConfigMap() string {
	return "observability-bundle-user-values"
}

// UnwireLogging unsets the user value confimap in a copy of the app
func (o Object) UnwireLogging(currentApp appv1.App) *appv1.App {
	desiredApp := currentApp.DeepCopy()

	observabilityBundleConfigMapMeta := common.ObservabilityBundleConfigMapMeta(&o)
	if desiredApp.Spec.UserConfig.ConfigMap.Name == observabilityBundleConfigMapMeta.GetName() ||
		desiredApp.Spec.UserConfig.ConfigMap.Namespace == observabilityBundleConfigMapMeta.GetNamespace() {
		desiredApp.Spec.UserConfig.ConfigMap.Name = ""
		desiredApp.Spec.UserConfig.ConfigMap.Namespace = ""
	}

	return desiredApp
}

// WireLogging sets the user value confimap in a copy of the app.
func (o Object) WireLogging(currentApp appv1.App) *appv1.App {
	desiredApp := currentApp.DeepCopy()

	observabilityBundleConfigMapMeta := common.ObservabilityBundleConfigMapMeta(&o)
	if desiredApp.Spec.UserConfig.ConfigMap.Name != observabilityBundleConfigMapMeta.GetName() ||
		desiredApp.Spec.UserConfig.ConfigMap.Namespace != observabilityBundleConfigMapMeta.GetNamespace() {
		desiredApp.Spec.UserConfig.ConfigMap.Name = observabilityBundleConfigMapMeta.GetName()
		desiredApp.Spec.UserConfig.ConfigMap.Namespace = observabilityBundleConfigMapMeta.GetNamespace()
	}

	return desiredApp
}
