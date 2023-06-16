package clusterhelpers

import (
	common "github.com/giantswarm/logging-operator/pkg/common"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func IsLoggingEnabled(object client.Object) bool {
	labels := object.GetLabels()

	return common.IsLoggingEnabled(labels)
}
