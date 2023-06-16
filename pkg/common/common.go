package common

import "github.com/giantswarm/logging-operator/pkg/key"

func IsLoggingEnabled(labels map[string]string) bool {
	value, ok := labels[key.LoggingLabel]

	return ok && value == "true"
}
