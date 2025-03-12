package capicluster

import (
	"fmt"
	"reflect"
	"strconv"

	appv1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/logging-operator/pkg/common"
	"github.com/giantswarm/logging-operator/pkg/key"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

type Object struct {
	client.Object
	Options loggedcluster.Options
}

func (o Object) HasLoggingEnabled() bool {
	labels := o.Object.GetLabels()

	// If logging is disabled at the installation level, we return false
	if !o.Options.EnableLoggingFlag {
		return false
	}

	loggingLabelValue, ok := labels[key.LoggingLabel]
	if !ok {
		return loggedcluster.LoggingEnabledDefault
	}

	loggingEnabled, err := strconv.ParseBool(loggingLabelValue)
	if err != nil {
		return loggedcluster.LoggingEnabledDefault
	}
	return loggingEnabled
}

func (o Object) GetLoggingAgent() string {
	return o.Options.LoggingAgent
}

func (o *Object) SetLoggingAgent(loggingAgent string) {
	o.Options.LoggingAgent = loggingAgent
}

func (o *Object) GetKubeEventsLogger() string {
	return o.Options.KubeEventsLogger
}

func (o *Object) SetKubeEventsLogger(kubeEventsLogger string) {
	o.Options.KubeEventsLogger = kubeEventsLogger
}

func (o Object) IsInsecureCA() bool {
	return o.Options.InsecureCA
}

func (o Object) GetAppsNamespace() string {
	return o.Object.GetNamespace()
}

func (o Object) AppConfigName(app string) string {
	return fmt.Sprintf("%s-%s", o.GetName(), app)
}

func (o Object) GetClusterName() string {
	return o.Object.GetName()
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
	return common.DefaultWriteTenant
}

func (o Object) IsCAPI() bool {
	return true
}

// On capi clusters, use an extraconfig
func (o Object) GetObservabilityBundleConfigMap() string {
	return "observability-bundle-logging-extraconfig"
}

func (o Object) getWiredExtraConfig() appv1.AppExtraConfig {
	observabilityBundleConfigMapMeta := common.ObservabilityBundleConfigMapMeta(&o)
	return appv1.AppExtraConfig{
		Kind:      "configMap",
		Name:      observabilityBundleConfigMapMeta.GetName(),
		Namespace: observabilityBundleConfigMapMeta.GetNamespace(),
		Priority:  25,
	}
}

// UnwireLogging unsets the extraconfig confimap in a copy of the app
func (o Object) UnwireLogging(currentApp appv1.App) *appv1.App {
	desiredApp := currentApp.DeepCopy()

	wiredExtraConfig := o.getWiredExtraConfig()
	for index, extraConfig := range currentApp.Spec.ExtraConfigs {
		if reflect.DeepEqual(extraConfig, wiredExtraConfig) {
			desiredApp.Spec.ExtraConfigs = append(currentApp.Spec.ExtraConfigs[:index], currentApp.Spec.ExtraConfigs[index+1:]...)
		}
	}

	return desiredApp
}

// WireLogging sets the extraconfig confimap in a copy of the app.
func (o Object) WireLogging(currentApp appv1.App) *appv1.App {
	desiredApp := currentApp.DeepCopy()
	wiredExtraConfig := o.getWiredExtraConfig()

	// We check if the extra config already exists to know if we need to remove it.
	var containsWiredExtraConfig bool = false
	for _, extraConfig := range currentApp.Spec.ExtraConfigs {
		if reflect.DeepEqual(extraConfig, wiredExtraConfig) {
			containsWiredExtraConfig = true
		}
	}

	if !containsWiredExtraConfig {
		desiredApp.Spec.ExtraConfigs = append(desiredApp.Spec.ExtraConfigs, wiredExtraConfig)
	}

	return desiredApp
}
