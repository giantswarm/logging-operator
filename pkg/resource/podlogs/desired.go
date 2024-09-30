package podlogs

import (
	"github.com/giantswarm/k8smetadata/pkg/label"
	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/logging-operator/pkg/key"
	podlogsv1alpha2 "github.com/giantswarm/logging-operator/pkg/resource/podlogs/apis/monitoring/v1alpha2"
)

const namespace = "giantswarm"

var replacement = "$1"

func PodLogs() []*PodLogsGetter {
	byPod := PodLogsGetter{
		podlogsv1alpha2.PodLogs{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "by-pod",
				Namespace: namespace,
				Labels: map[string]string{
					label.ManagedBy: "logging-operator",
				},
			},
			Spec: podlogsv1alpha2.PodLogsSpec{
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
						Replacement:  &replacement,
						Regex:        "(.*)",
					},
				},
			},
		},
	}

	byNamespace := PodLogsGetter{
		podlogsv1alpha2.PodLogs{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "by-namespace",
				Namespace: namespace,
				Labels: map[string]string{
					label.ManagedBy: "logging-operator",
				},
			},
			Spec: podlogsv1alpha2.PodLogsSpec{
				Selector: metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      key.LoggingLabel,
							Operator: metav1.LabelSelectorOpNotIn,
							Values:   []string{"disabled"},
						},
					},
				},
				NamespaceSelector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						key.LoggingLabel: "enabled",
					},
				},
				RelabelConfigs: []*promv1.RelabelConfig{
					{
						Action:       "replace",
						SourceLabels: []promv1.LabelName{"__meta_kubernetes_namespace_label_namespace_giantswarm_io_logging_tenant"},
						TargetLabel:  "tenant_id",
						Replacement:  &replacement,
						Regex:        "(.*)",
					},
				},
			},
		},
	}

	podlogs := []*PodLogsGetter{
		&byPod,
		&byNamespace,
	}

	return podlogs
}
