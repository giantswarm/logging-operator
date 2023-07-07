package common

import (
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

func IsLoggingEnabled(lc loggedcluster.Interface) bool {

	// Logging should be enabled when all conditions are met:
	//   - logging label is set and true on the cluster
	//   - cluster is not being deleted
	//   - TODO(theo) global logging flag is enabled

	return lc.GetLoggingLabel() == "true" && lc.GetDeletionTimestamp().IsZero() && lc.GetEnableLoggingFlag()
}

func AddCommonLabels(labels map[string]string) {
	labels["giantswarm.io/managed-by"] = "logging-operator"
}
