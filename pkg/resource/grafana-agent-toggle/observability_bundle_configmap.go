package grafanaagenttoggle

import (
	"github.com/blang/semver"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

type Values struct {
	Apps map[string]app `yaml:"apps" json:"apps"`
}

type app struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// GenerateObservabilityBundleConfigMap returns a configmap for
// the observabilitybundle application to enable grafana-agent.
func GenerateObservabilityBundleConfigMap(lc loggedcluster.Interface, observabilityBundleVersion semver.Version) (v1.ConfigMap, error) {
	appName := "grafanaAgent"
	if observabilityBundleVersion.LT(semver.MustParse("1.0.0")) {
		appName = "grafanaAgent"
	}

	values := Values{
		Apps: map[string]app{
			appName: {
				Enabled: true,
			},
		},
	}

	v, err := yaml.Marshal(values)
	if err != nil {
		return v1.ConfigMap{}, errors.WithStack(err)
	}

	configmap := v1.ConfigMap{
		ObjectMeta: common.ObservabilityBundleConfigMapMeta(lc),
		Data: map[string]string{
			"values": string(v),
		},
	}

	return configmap, nil
}
