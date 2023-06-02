package clusterhelpers

import (
	common "github.com/giantswarm/logging-operator/pkg/common"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

func IsLoggingEnabled(cluster capiv1beta1.Cluster) bool {
	labels := cluster.GetLabels()

	return common.IsLoggingEnabled(labels)
}
