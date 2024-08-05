package loggingagentstoggle

import (
	"context"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

type Values struct {
	Apps map[string]app `yaml:"apps" json:"apps"`
}

type app struct {
	Enabled   bool   `yaml:"enabled" json:"enabled"`
	Namespace string `yaml:"namespace,omitempty" json:"namespace,omitempty"`
}

// GenerateObservabilityBundleConfigMap returns a configmap for
// the observabilitybundle application to enable logging agents.
func GenerateObservabilityBundleConfigMap(ctx context.Context, lc loggedcluster.Interface, observabilityBundleVersion semver.Version) (v1.ConfigMap, error) {
	appsToEnable := map[string]app{}

	promtailAppName := "promtail"
	if observabilityBundleVersion.LT(semver.MustParse("1.0.0")) {
		promtailAppName = "promtail-app"
	}

	// Enforce promtail as logging agent when observability-bundle version < 1.5.0
	if observabilityBundleVersion.LT(semver.MustParse("1.5.0")) && lc.GetLoggingAgent() == common.LoggingAgentAlloy {
		logger := log.FromContext(ctx)
		logger.Info("Logging agent is not supported by observability bundle, using promtail instead.", "observability-bundle-version", observabilityBundleVersion, "logging-agent", lc.GetLoggingAgent())
		lc.SetLoggingAgent(common.LoggingAgentPromtail)
	}

	switch lc.GetLoggingAgent() {
	case common.LoggingAgentPromtail:
		appsToEnable[promtailAppName] = app{
			Enabled: true,
		}
		appsToEnable[common.AlloyLogAgentAppName] = app{
			Enabled: false,
		}
	case common.LoggingAgentAlloy:
		appsToEnable[common.AlloyLogAgentAppName] = app{
			Enabled:   true,
			Namespace: common.AlloyLogAgentAppNamespace,
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
