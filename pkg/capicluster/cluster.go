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
		return true
	}

	loggingEnabled, err := strconv.ParseBool(loggingLabelValue)
	if err != nil {
		return false
	}
	return loggingEnabled
}

func (o Object) GetAppsNamespace() string {
	return o.Object.GetNamespace()
}

func (o Object) AppConfigName(app string) string {
	return fmt.Sprintf("%s-%s", o.GetName(), app)
}

func (o Object) ObservabilityBundleConfigLabelName(config string) string {
	return config
}

func (o Object) GetClusterName() string {
	return o.Object.GetName()
}

func (o Object) GetInstallationName() string {
	return o.Options.InstallationName
}

func (o Object) GetRegion() string {
	return o.Options.InstallationRegion
}

func (o Object) GetCloudDomain() string {
	return o.Options.InstallationBaseDomain
}

func (o Object) GetEnableLoggingFlag() bool {
	return o.Options.EnableLoggingFlag
}

func (o Object) GetObject() client.Object {
	return o.Object
}

// On capi clusters, use an extraconfig
func (o Object) GetObservabilityBundleConfigMap() string {
	return "observability-bundle-logging-extraconfig"
}

func (o Object) getWiredExtraConfig() appv1.AppExtraConfig {
	observabilityBundleConfigMapMeta := common.ObservabilityBundleConfigMapMeta(o)
	return appv1.AppExtraConfig{
		Kind:      "configMap",
		Name:      observabilityBundleConfigMapMeta.GetName(),
		Namespace: observabilityBundleConfigMapMeta.GetNamespace(),
		Priority:  25,
	}
}

// UnwirePromtail unsets the extraconfig confimap in a copy of the app
func (o Object) UnwirePromtail(currentApp appv1.App) *appv1.App {
	desiredApp := currentApp.DeepCopy()

	wiredExtraConfig := o.getWiredExtraConfig()
	for index, extraConfig := range currentApp.Spec.ExtraConfigs {
		if reflect.DeepEqual(extraConfig, wiredExtraConfig) {
			desiredApp.Spec.ExtraConfigs = append(currentApp.Spec.ExtraConfigs[:index], currentApp.Spec.ExtraConfigs[index+1:]...)
		}
	}

	return desiredApp
}

// WirePromtail sets the extraconfig confimap in a copy of the app.
func (o Object) WirePromtail(currentApp appv1.App) *appv1.App {
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
