package common

func IsLoggingEnabled(labels map[string]string) bool {
	value, ok := labels["giantswarm.io/logging"]

	return ok && value == "true"
}
