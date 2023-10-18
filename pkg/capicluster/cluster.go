package capicluster

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/logging-operator/pkg/key"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

type Object struct {
	client.Object
	Options loggedcluster.Options
}

func (o Object) GetLoggingLabel() string {
	labels := o.Object.GetLabels()

	value := labels[key.LoggingLabel]

	return value
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

func (o Object) IsVintage() bool {
	return false
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

// On capi clusters, use an extraconfig
func (o Object) GetObservabilityBundleConfigMap() string {
	return fmt.Sprintf("%s-observability-bundle-logging-extraconfig", o.GetClusterName())
}
