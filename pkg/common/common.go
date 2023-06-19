package common

import (
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

func IsLoggingEnabled(object loggedcluster.Interface) bool {

	return object.GetLoggingLabel() == "true" && object.GetDeletionTimestamp().IsZero()
}
