package podlogs

import (
	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/logging-operator/pkg/key"
	podlogsv1alpha2 "github.com/giantswarm/logging-operator/pkg/resource/podlogs/apis/monitoring/v1alpha2"
)

func PodLogs() *podlogsv1alpha2.PodLogs {
	p := podlogsv1alpha2.PodLogs{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "by-pod",
			Namespace: "giantswarm",
			Labels: map[string]string{
				"giantswarm.io/managed-by": "logging-operator",
			},
		},
	}

	return &p
}

func PodLogsSpec() podlogsv1alpha2.PodLogsSpec {
	p := podlogsv1alpha2.PodLogsSpec{
		Selector: metav1.LabelSelector{
			MatchLabels: map[string]string{
				key.LoggingLabel: "enabled",
			},
		},
		NamespaceSelector: metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      key.LoggingLabel,
					Operator: metav1.LabelSelectorOpNotIn,
					Values:   []string{"enabled"},
				},
			},
		},
		RelabelConfigs: []*promv1.RelabelConfig{
			{
				Action:       "replace",
				SourceLabels: []promv1.LabelName{"__meta_kubernetes_pod_label_giantswarm_io_logging_tenant"},
				TargetLabel:  "tenant_id",
				Regex:        "(.*)",
			},
		},
	}

	return p
}
