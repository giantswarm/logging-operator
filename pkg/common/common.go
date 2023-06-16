package common

import (
	"github.com/giantswarm/logging-operator/pkg/key"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func IsLoggingEnabled(object client.Object) bool {
	labels := object.GetLabels()
	value, ok := labels[key.LoggingLabel]

	return ok && value == "true"
}
