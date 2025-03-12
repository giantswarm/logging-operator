package controller

import (
	"context"

	"github.com/blang/semver"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

var (
	supportAlloyEvents = semver.MustParse("1.9.0")
	supportAlloyLogs   = semver.MustParse("1.6.0")
)

func toggleAgents(ctx context.Context, client client.Client, lc loggedcluster.Interface) error {
	logger := log.FromContext(ctx)

	observabilityBundleVersion, err := common.GetObservabilityBundleAppVersion(ctx, client, lc)
	if err != nil {
		return err
	}

	// Enforce promtail as logging agent when observability-bundle version < 1.6.0 because this needs alloy 0.4.0.
	if observabilityBundleVersion.LT(supportAlloyLogs) && lc.GetLoggingAgent() == common.LoggingAgentAlloy {
		logger.Info("Alloy logging agent is not supported by observability bundle, using promtail instead.", "observability-bundle-version", observabilityBundleVersion, "logging-agent", lc.GetLoggingAgent())
		lc.SetLoggingAgent(common.LoggingAgentPromtail)
	}

	// Enforce grafana-agent as events logger when observability-bundle version < 1.9.0 because this needs alloy 0.7.0.
	if observabilityBundleVersion.LT(supportAlloyEvents) && lc.GetKubeEventsLogger() == common.EventsLoggerAlloy {
		logger.Info("Alloy events logger is not supported by observability bundle, using grafana-agent instead.", "observability-bundle-version", observabilityBundleVersion, "events-logger", lc.GetKubeEventsLogger())
		lc.SetKubeEventsLogger(common.EventsLoggerGrafanaAgent)
	}

	return nil
}
