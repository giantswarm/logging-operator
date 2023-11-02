package common

import (
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

// ReadUser is the global user for reading logs
const ReadUser = "read"

func IsLoggingEnabled(lc loggedcluster.Interface) bool {
	// Logging should be enabled when all conditions are met:
	//   - logging label is set and true on the cluster
	//   - cluster is not being deleted
	//   - global logging flag is enabled
	return lc.HasLoggingEnabled() && lc.GetDeletionTimestamp().IsZero() && lc.GetEnableLoggingFlag()
}

func AddCommonLabels(labels map[string]string) {
	labels["giantswarm.io/managed-by"] = "logging-operator"
}

func IsWorkloadCluster(lc loggedcluster.Interface) bool {
	return lc.GetInstallationName() != lc.GetClusterName()
}
