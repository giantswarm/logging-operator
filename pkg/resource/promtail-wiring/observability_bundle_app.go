package promtailwiring

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ObservabilityBundleAppMeta returns metadata for the observability bundle app.
func ObservabilityBundleAppMeta(object client.Object) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      fmt.Sprintf("%s-observability-bundle", object.GetName()),
		Namespace: object.GetName(),
	}
}
