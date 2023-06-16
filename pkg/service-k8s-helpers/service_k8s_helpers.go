package servicek8shelpers

import (
	common "github.com/giantswarm/logging-operator/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

func IsLoggingEnabled(serviceK8S corev1.Service) bool {
	labels := serviceK8S.GetLabels()

	return common.IsLoggingEnabled(labels)
}
