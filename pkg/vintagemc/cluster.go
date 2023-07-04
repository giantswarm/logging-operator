package vintagemc

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Object struct {
	client.Object
}

func (o Object) GetLoggingLabel() string {
	return "true"
}

func (o Object) GetAppsNamespace() string {
	return "giantswarm"
}

func (o Object) AppConfigName(app string) string {
	return app
}

func (o Object) GetClusterName() string {
	// TODO - return installation name for vintage MC
	return "gauss"
}

func (o Object) GetObject() client.Object {
	return o.Object
}

// On vintage MC, there's no support for extraconfig so we should use standard user values
func (o Object) GetObservabilityBundleConfigMap() string {
	return "observability-bundle-user-values"
}
