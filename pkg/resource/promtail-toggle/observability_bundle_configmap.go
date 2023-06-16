package promtailtoggle

import (
	"fmt"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/yaml"
)

type Values struct {
	Apps map[string]app `yaml:"apps" json:"apps"`
}

type app struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

func ObservabilityBundleConfigMapMeta(cluster capiv1beta1.Cluster) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      fmt.Sprintf("%s-observability-bundle-user-values", cluster.GetName()),
		Namespace: cluster.GetName(),
	}
}

func GenerateObservabilityBundleConfigMap(cluster capiv1beta1.Cluster) (v1.ConfigMap, error) {
	values := Values{
		Apps: map[string]app{
			"promtail-app": {
				Enabled: true,
			},
		},
	}

	v, err := yaml.Marshal(values)
	if err != nil {
		return v1.ConfigMap{}, errors.WithStack(err)
	}

	configmap := v1.ConfigMap{
		ObjectMeta: ObservabilityBundleConfigMapMeta(cluster),
		Data: map[string]string{
			"values": string(v),
		},
	}

	return configmap, nil
}
