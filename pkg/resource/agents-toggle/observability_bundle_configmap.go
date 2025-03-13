package agentstoggle

import (
	"context"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

var (
	supportAlloyEvents = semver.MustParse("1.9.0")
	supportAlloyLogs   = semver.MustParse("1.6.0")
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
func generateObservabilityBundleConfig(ctx context.Context, lc loggedcluster.Interface, observabilityBundleVersion semver.Version) (string, error) {
	appsToEnable := map[string]app{}

	if err := toggleLogAgent(ctx, lc, observabilityBundleVersion, appsToEnable); err != nil {
		return "", errors.WithStack(err)
	}

	if err := toggleKubeEventsLogger(ctx, lc, observabilityBundleVersion, appsToEnable); err != nil {
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
func toggleLogAgent(ctx context.Context, lc loggedcluster.Interface, observabilityBundleVersion semver.Version, appsToEnable map[string]app) error {
	logger := log.FromContext(ctx)
	// Enforce promtail as logging agent when observability-bundle version < 1.6.0 because this needs alloy 0.4.0.
	if observabilityBundleVersion.LT(supportAlloyLogs) && lc.GetLoggingAgent() == common.LoggingAgentAlloy {
		logger.Info("Alloy logging agent is not supported by observability bundle, using promtail instead.", "observability-bundle-version", observabilityBundleVersion, "logging-agent", lc.GetLoggingAgent())
		lc.SetLoggingAgent(common.LoggingAgentPromtail)
	}

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
func toggleKubeEventsLogger(ctx context.Context, lc loggedcluster.Interface, observabilityBundleVersion semver.Version, appsToEnable map[string]app) error {
	logger := log.FromContext(ctx)

	// Enforce grafana-agent as events logger when observability-bundle version < 1.9.0 because this needs alloy 0.7.0.
	if observabilityBundleVersion.LT(supportAlloyEvents) && lc.GetKubeEventsLogger() == common.EventsLoggerAlloy {
		logger.Info("Alloy events logger is not supported by observability bundle, using grafana-agent instead.", "observability-bundle-version", observabilityBundleVersion, "events-logger", lc.GetKubeEventsLogger())
		lc.SetKubeEventsLogger(common.EventsLoggerGrafanaAgent)
	}

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
