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
	kubeSystem := PodLogsGetter{
		podlogsv1alpha2.PodLogs{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kube-system",
				Namespace: namespace,
				Labels: map[string]string{
					label.ManagedBy: "logging-operator",
				},
			},
			Spec: podlogsv1alpha2.PodLogsSpec{
				NamespaceSelector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						"kubernetes.io/metadata.name": "kube-system",
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

	podSelector := PodLogsGetter{
		podlogsv1alpha2.PodLogs{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod-selector",
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
							Operator: metav1.LabelSelectorOpDoesNotExist,
						},
						{
							Key:      "kubernetes.io/metadata.name",
							Operator: metav1.LabelSelectorOpNotIn,
							Values:   []string{"kube-system"},
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

	namespaceSelector := PodLogsGetter{
		podlogsv1alpha2.PodLogs{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "namespace-selector",
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
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      key.LoggingLabel,
							Operator: metav1.LabelSelectorOpIn,
							Values:   []string{"enabled"},
						},
						{
							Key:      "kubernetes.io/metadata.name",
							Operator: metav1.LabelSelectorOpNotIn,
							Values:   []string{"kube-system"},
						},
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
		&kubeSystem,
		&podSelector,
		&namespaceSelector,
	}

	return podlogs
}
