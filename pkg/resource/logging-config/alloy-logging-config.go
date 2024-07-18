package loggingconfig

import (
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

// GenerateAlloyLoggingConfig returns a configmap for
// the logging extra-config
// This is a plaholder function, in case we need to populate the configmap
// for Alloy in the future.
func GenerateAlloyLoggingConfig(lc loggedcluster.Interface) (string, error) {
	return "", nil
}
