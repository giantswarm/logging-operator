package loggingagentstoggle

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
// the observabilitybundle application to enable logging agents.
func GenerateObservabilityBundleConfigMap(lc loggedcluster.Interface, observabilityBundleVersion semver.Version) (v1.ConfigMap, error) {
	appsToEnable := map[string]app{}

	promtailAppName := "promtail"
	if observabilityBundleVersion.LT(semver.MustParse("1.0.0")) {
		promtailAppName = "promtail-app"
	}

	switch lc.GetLoggingAgent() {
	case "promtail":
		appsToEnable[promtailAppName] = app{
			Enabled: true,
		}
		appsToEnable["alloy-logs"] = app{
			Enabled: false,
		}
	case "alloy":
		appsToEnable["alloy-logs"] = app{
			Enabled: true,
		}
		appsToEnable[promtailAppName] = app{
			Enabled: false,
		}
	default:
		return v1.ConfigMap{}, errors.Errorf("unsupported logging agent %q", lc.GetLoggingAgent())
	}

	if observabilityBundleVersion.GE(semver.MustParse("0.10.0")) {
		appsToEnable["grafanaAgent"] = app{
			Enabled: true,
		}
	}
	values := Values{
		Apps: appsToEnable,
	}

	v, err := yaml.Marshal(values)
	if err != nil {
		return v1.ConfigMap{}, errors.WithStack(err)
	}

	configmap := v1.ConfigMap{
		ObjectMeta: common.ObservabilityBundleConfigMapMeta(lc),
		Data:       map[string]string{"values": string(v)},
	}

	return configmap, nil
}
