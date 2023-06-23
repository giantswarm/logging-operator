package capicluster

import (
	"fmt"

	"github.com/giantswarm/logging-operator/pkg/key"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	return o.Object.GetNamespace()
}

func (o Object) AppConfigName(app string) string {
	return fmt.Sprintf("%s-%s", o.GetName(), app)
}

func (o Object) GetObject() client.Object {
	return o.Object
}
