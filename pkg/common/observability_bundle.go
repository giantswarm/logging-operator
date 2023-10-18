package common

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

const ObservabilityBundleAppName string = "observability-bundle"

// ObservabilityBundleAppMeta returns metadata for the observability bundle app.
func ObservabilityBundleAppMeta(lc loggedcluster.Interface) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      lc.AppConfigName(ObservabilityBundleAppName),
		Namespace: lc.GetAppsNamespace(),
		Labels:    map[string]string{},
	}

	AddCommonLabels(metadata.Labels)
	return metadata
}

// ObservabilityBundleConfigMapMeta returns metadata for the observability bundle extra values configmap.
func ObservabilityBundleConfigMapMeta(lc loggedcluster.Interface) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      lc.AppConfigName(lc.GetObservabilityBundleConfigMap()),
		Namespace: lc.GetAppsNamespace(),
		Labels: map[string]string{
			// This label is used by cluster-operator to find extraconfig. This only works on vintage WCs
			"app.kubernetes.io/name": lc.ObservabilityBundleConfigLabelName(ObservabilityBundleAppName),
		},
	}

	AddCommonLabels(metadata.Labels)
	return metadata
}
