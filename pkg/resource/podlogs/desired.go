package podlogs

import (
	"github.com/giantswarm/k8smetadata/pkg/label"
	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	podlogsv1alpha2 "github.com/giantswarm/logging-operator/pkg/resource/podlogs/apis/monitoring/v1alpha2"
)

func PodLog() *PodLogsGetter {
	podLog := &PodLogsGetter{
		podlogsv1alpha2.PodLogs{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "logging-operator",
				Namespace: "giantswarm",
				Labels: map[string]string{
					label.ManagedBy: "logging-operator",
				},
			},
			Spec: podlogsv1alpha2.PodLogsSpec{
				// Select all pods
				Selector: metav1.LabelSelector{},
				// from every namespace
				NamespaceSelector: metav1.LabelSelector{},
				RelabelConfigs: []*promv1.RelabelConfig{
					// Drop pod with explicit logging disabled
					{
						Action: "drop",
						Regex:  ".*disabled.*",
						SourceLabels: []promv1.LabelName{
							"__meta_kubernetes_namespace_label_giantswarm_io_logging",
							"__meta_kubernetes_pod_label_giantswarm_io_logging",
						},
					},
					// Only include pod with explicit logging enabled
					{
						Action: "keep",
						Regex:  ".*enabled.*",
						SourceLabels: []promv1.LabelName{
							"__meta_kubernetes_namespace_label_giantswarm_io_logging",
							"__meta_kubernetes_pod_label_giantswarm_io_logging",
						},
					},
				},
			},
		},
	}

	return podLog
}
