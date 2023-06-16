package promtailwiring

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// ObservabilityBundleAppMeta returns metadata for the observability bundle app.
func ObservabilityBundleAppMeta(cluster capiv1beta1.Cluster) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      fmt.Sprintf("%s-observability-bundle", cluster.GetName()),
		Namespace: cluster.GetName(),
	}
}
