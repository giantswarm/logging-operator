package common

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/blang/semver"
	appv1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"

	"github.com/giantswarm/logging-operator/pkg/config"
)

var (
	supportAlloyEvents = semver.MustParse("1.9.0")
	supportAlloyLogs   = semver.MustParse("1.6.0")
)

const (
	observabilityBundleConfigMapName        = "observability-bundle-logging-extraconfig"
	observabilityBundleAppName       string = "observability-bundle"
)

// ObservabilityBundleAppMeta returns metadata for the observability bundle app.
func ObservabilityBundleAppMeta(cluster *capi.Cluster) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      AppConfigName(cluster, observabilityBundleAppName),
		Namespace: cluster.GetNamespace(),
		Labels:    map[string]string{},
	}

	AddCommonLabels(metadata.Labels)
	return metadata
}

// ObservabilityBundleConfigMapMeta returns metadata for the observability bundle extra values configmap.
func ObservabilityBundleConfigMapMeta(cluster *capi.Cluster) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      AppConfigName(cluster, observabilityBundleConfigMapName),
		Namespace: cluster.GetNamespace(),
		Labels: map[string]string{
			// This label is used by cluster-operator to find extraconfig. This only works on vintage WCs
			"app.kubernetes.io/name": observabilityBundleAppName,
		},
	}

	AddCommonLabels(metadata.Labels)
	return metadata
}

func GetObservabilityBundleAppVersion(ctx context.Context, client client.Client, cluster *capi.Cluster) (version semver.Version, err error) {
	// Get observability bundle app metadata.
	appMeta := ObservabilityBundleAppMeta(cluster)
	// Retrieve the app.
	var currentApp appv1.App
	err = client.Get(ctx, types.NamespacedName{Name: appMeta.GetName(), Namespace: appMeta.GetNamespace()}, &currentApp)
	if err != nil {
		return version, err
	}
	return semver.Parse(currentApp.Spec.Version)
}

func ToggleAgents(ctx context.Context, client client.Client, cluster *capi.Cluster, cfg config.Config) (*LoggingAgent, error) {
	logger := log.FromContext(ctx)

	observabilityBundleVersion, err := GetObservabilityBundleAppVersion(ctx, client, cluster)
	if err != nil {
		return nil, err
	}
	agent := &LoggingAgent{
		LoggingAgent:     cfg.DefaultLoggingAgent,
		KubeEventsLogger: cfg.DefaultKubeEventsLogger,
	}

	// Enforce promtail as logging agent when observability-bundle version < 1.6.0 because this needs alloy 0.4.0.
	if observabilityBundleVersion.LT(supportAlloyLogs) && agent.GetLoggingAgent() == LoggingAgentAlloy {
		logger.Info("Alloy logging agent is not supported by observability bundle, using promtail instead.", "observability-bundle-version", observabilityBundleVersion, "logging-agent", agent.GetLoggingAgent())
		agent.SetLoggingAgent(LoggingAgentPromtail)
	}

	// Enforce grafana-agent as events logger when observability-bundle version < 1.9.0 because this needs alloy 0.7.0.
	if observabilityBundleVersion.LT(supportAlloyEvents) && agent.GetKubeEventsLogger() == EventsLoggerAlloy {
		logger.Info("Alloy events logger is not supported by observability bundle, using grafana-agent instead.", "observability-bundle-version", observabilityBundleVersion, "events-logger", agent.GetKubeEventsLogger())
		agent.SetKubeEventsLogger(EventsLoggerGrafanaAgent)
	}

	return agent, nil
}
