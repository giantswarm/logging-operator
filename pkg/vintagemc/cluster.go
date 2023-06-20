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
