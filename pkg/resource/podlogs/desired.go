package podlogs

import (
	_ "embed"

	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/logging-operator/pkg/key"
	podlogsv1alpha2 "github.com/giantswarm/logging-operator/pkg/resource/podlogs/apis/monitoring/v1alpha2"
)

var replacement = "$1"

type PodLogsObject struct {
	podlogsv1alpha2.PodLogs
}

func (p *PodLogsObject) GetBase() *podlogsv1alpha2.PodLogs {
	pl := podlogsv1alpha2.PodLogs{
		ObjectMeta: p.ObjectMeta,
	}

	return &pl
}

func (p *PodLogsObject) GetSpec() podlogsv1alpha2.PodLogsSpec {
	return p.Spec
}

var (
	//go:embed resources/podlogs_pod.yaml
	byPodContent []byte
	byPodLog     *podlogsv1alpha2.PodLogs
)

func init() {
	err := yaml.Unmarshal(byPodContent, &byPodLog)
	if err != nil {
		panic(err)
	}
}

func PodLogs() []*PodLogsObject {
	byPod := PodLogsObject{
		*byPodLog.DeepCopy(),
	}

	byNamespace := PodLogsObject{
		podlogsv1alpha2.PodLogs{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "by-namespace",
				Namespace: "giantswarm",
				Labels: map[string]string{
					"giantswarm.io/managed-by": "logging-operator",
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

	items := []*PodLogsObject{
		&byPod,
		&byNamespace,
	}

	return items
}
