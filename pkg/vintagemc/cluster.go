package vintagemc

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

type Object struct {
	client.Object
	Options loggedcluster.Options
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
	// return installation name for vintage MC
	return o.Options.InstallationName
}

func (o Object) GetInstallationName() string {
	return o.Options.InstallationName
}

func (o Object) GetDisableLoggingFlag() bool {
	return o.Options.DisableLoggingFlag
}

func (o Object) GetObject() client.Object {
	return o.Object
}

// On vintage MC, there's no support for extraconfig so we should use standard user values
func (o Object) GetObservabilityBundleConfigMap() string {
	return "observability-bundle-user-values"
}
