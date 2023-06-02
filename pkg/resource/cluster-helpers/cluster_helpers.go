package clusterhelpers

import (
	"github.com/giantswarm/logging-operator/pkg/key"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

func IsLoggingEnabled(cluster capiv1beta1.Cluster) bool {
	labels := cluster.GetLabels()

	value, ok := labels[key.LoggingLabel]

	return ok && value == "true"
}
