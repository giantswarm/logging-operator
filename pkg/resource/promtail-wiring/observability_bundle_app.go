package promtailwiring

import (
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ObservabilityBundleAppMeta returns metadata for the observability bundle app.
func ObservabilityBundleAppMeta(lc loggedcluster.Interface) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      lc.AppConfigName("observability-bundle"),
		Namespace: lc.GetAppsNamespace(),
	}
}
