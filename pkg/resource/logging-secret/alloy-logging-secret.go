package loggingsecret

import (
	v1 "k8s.io/api/core/v1"

	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

func GenerateAlloyLoggingSecret(lc loggedcluster.Interface, credentialsSecret *v1.Secret, lokiURL string) (map[string][]byte, error) {
	data := make(map[string][]byte)

	return data, nil
}
