package common

import (
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

func IsLoggingEnabled(lc loggedcluster.Interface) bool {

	return lc.GetLoggingLabel() == "true" && lc.GetDeletionTimestamp().IsZero()
}
