package promtailtoggle

import (
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type Values struct {
	Apps map[string]app `yaml:"apps" json:"apps"`
}

type app struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// ObservabilityBundleConfigMapMeta returns metadata for the observability bundle user value configmap.
func ObservabilityBundleConfigMapMeta(lc loggedcluster.Interface) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      lc.GetAppName("observability-bundle-user-values"),
		Namespace: lc.GetAppsNamespace(),
	}
}

// GenerateObservabilityBundleConfigMap returns a configmap for
// the observabilitybundle application to enable promtail.
func GenerateObservabilityBundleConfigMap(lc loggedcluster.Interface) (v1.ConfigMap, error) {
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
		ObjectMeta: ObservabilityBundleConfigMapMeta(lc),
		Data: map[string]string{
			"values": string(v),
		},
	}

	return configmap, nil
}
