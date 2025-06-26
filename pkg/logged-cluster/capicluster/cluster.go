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
	*loggedcluster.LoggingAgent
}

func (o Object) HasLoggingEnabled(enableLoggingFlag bool) bool {
	labels := o.GetLabels()

	// If logging is disabled at the installation level, we return false
	if !enableLoggingFlag {
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

func (o Object) GetAppsNamespace() string {
	return o.GetNamespace()
}

func (o Object) AppConfigName(app string) string {
	return fmt.Sprintf("%s-%s", o.GetName(), app)
}

func (o Object) GetClusterName() string {
	return o.GetName()
}

func (o Object) GetObject() client.Object {
	return o.Object
}

func (o Object) GetTenant() string {
	return common.DefaultWriteTenant
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
	var containsWiredExtraConfig = false
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
