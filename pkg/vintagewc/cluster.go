package vintagewc

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/logging-operator/pkg/key"
)

type Object struct {
	client.Object
}

func (o Object) GetLoggingLabel() string {
	labels := o.Object.GetLabels()

	value := labels[key.LoggingLabel]

	return value
}

func (o Object) GetAppsNamespace() string {
	return o.Object.GetName()
}

func (o Object) AppConfigName(app string) string {
	return app
}

func (o Object) GetObject() client.Object {
	return o.Object
}

// on vintage WC, use extraconfig
func (o Object) GetObservabilityBundleConfigMap() string {
	return "observability-bundle-logging-values"
}