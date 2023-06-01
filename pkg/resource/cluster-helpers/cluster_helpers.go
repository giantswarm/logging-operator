package clusterhelpers

import capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"

func IsLoggingEnabled(cluster capiv1beta1.Cluster) bool {
	labels := cluster.GetLabels()

	value, ok := labels["giantswarm.io/logging"]

	return ok && value == "true"
}
