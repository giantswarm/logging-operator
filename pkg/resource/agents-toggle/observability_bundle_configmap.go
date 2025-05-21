package agentstoggle

import (
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

type values struct {
	Apps map[string]app `yaml:"apps" json:"apps"`
}

type app struct {
	Enabled   bool   `yaml:"enabled" json:"enabled"`
	Namespace string `yaml:"namespace,omitempty" json:"namespace,omitempty"`
}

// generateObservabilityBundleConfig returns a configmap for
// the observabilitybundle application to enable logging agents and events-loggers.
func generateObservabilityBundleConfig(lc loggedcluster.Interface) (string, error) {
	appsToEnable := map[string]app{}

	if err := toggleLogAgent(lc, appsToEnable); err != nil {
		return "", errors.WithStack(err)
	}

	if err := toggleKubeEventsLogger(lc, appsToEnable); err != nil {
		return "", errors.WithStack(err)
	}

	values := values{
		Apps: appsToEnable,
	}

	v, err := yaml.Marshal(values)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return string(v), nil
}

// toggleLogAgent toggles the logging agent based on the observability bundle version.
func toggleLogAgent(lc loggedcluster.Interface, appsToEnable map[string]app) error {
	switch lc.GetLoggingAgent() {
	case common.LoggingAgentPromtail:
		appsToEnable[common.PromtailObservabilityBundleAppName] = app{
			Enabled: true,
		}
		appsToEnable[common.AlloyObservabilityBundleAppName] = app{
			Enabled: false,
		}
	case common.LoggingAgentAlloy:
		appsToEnable[common.AlloyObservabilityBundleAppName] = app{
			Enabled:   true,
			Namespace: common.AlloyLogAgentAppNamespace,
		}
		appsToEnable[common.PromtailObservabilityBundleAppName] = app{
			Enabled: false,
		}
	default:
		return errors.Errorf("unsupported logging agent %q", lc.GetLoggingAgent())
	}

	return nil
}

// toggleKubeEventsLogger toggles the kube-events-logger based on the observability bundle version.
func toggleKubeEventsLogger(lc loggedcluster.Interface, appsToEnable map[string]app) error {
	switch lc.GetKubeEventsLogger() {
	case common.EventsLoggerGrafanaAgent:
		appsToEnable["grafanaAgent"] = app{
			Enabled: true,
		}
		appsToEnable["alloyEvents"] = app{
			Enabled: false,
		}
	case common.EventsLoggerAlloy:
		appsToEnable["grafanaAgent"] = app{
			Enabled: false,
		}
		appsToEnable["alloyEvents"] = app{
			Enabled: true,
		}
	default:
		return errors.Errorf("unsupported events logger %q", lc.GetKubeEventsLogger())
	}

	return nil
}
