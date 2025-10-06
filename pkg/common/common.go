package common

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/logging-operator/pkg/key"
)

const (
	// LoggingEnabledDefault defines if WCs logs are collected by default
	LoggingEnabledDefault = true

	// ReadUser is the global user for reading logs
	ReadUser = "read"
	// DefaultWriteTenant is the default tenant for writing logs
	DefaultWriteTenant = "giantswarm"
	// Loki Gateway Ingress
	lokiGatewayIngressNamespace = "loki"
	lokiGatewayIngressName      = "loki-gateway"
	// Tempo Ingress used to send traces to Tempo
	// Ingress resources are defined here: https://github.com/giantswarm/tempo-app/blob/main/helm/tempo/templates/ingress.yaml
	// Configuration for ingresses are here: https://github.com/giantswarm/shared-configs/blob/main/default/apps/tempo/configmap-values.yaml.template#L144-L157
	tempoIngressNamespace = "tempo"
	tempoIngressName      = "tempo"
	// grafana-agent secret name
	//#nosec G101
	grafanaAgentExtraSecretName = "grafana-agent-secret"

	// Possible values for --logging-agent flag.
	LoggingAgentPromtail = "promtail"
	LoggingAgentAlloy    = "alloy"

	// Possible values for --events-logger flag.
	EventsLoggerAlloy        = "alloy"
	EventsLoggerGrafanaAgent = "grafana-agent"

	// App name keys in the observability bundle
	AlloyObservabilityBundleAppName    = "alloyLogs"
	PromtailObservabilityBundleAppName = "promtail"

	// Alloy app name and namespace when using Alloy as logging agent.
	AlloyLogAgentAppName      = "alloy-logs"
	AlloyLogAgentAppNamespace = "kube-system"

	PriorityClassName = "giantswarm-critical"

	// Alloy app name and namespace when using Alloy as events logger.
	AlloyEventsLoggerAppName      = "alloy-events"
	AlloyEventsLoggerAppNamespace = "kube-system"

	// LokiMaxBackoffPeriod specifies the maximum retry backoff duration for Loki writes.
	LokiMaxBackoffPeriod = 10 * time.Minute
	// LokiRemoteTimeout configures the write timeout for remote Loki endpoints.
	LokiRemoteTimeout = 60 * time.Second

	LokiBaseURLFormat = "https://%s"
	lokiAPIV1PushPath = "/loki/api/v1/push"
	LokiPushURLFormat = LokiBaseURLFormat + lokiAPIV1PushPath

	LoggingURL      = "logging-url"
	LoggingTenantID = "logging-tenant-id"
	LoggingUsername = "logging-username"
	LoggingPassword = "logging-password"
	LokiRulerAPIURL = "ruler-api-url"
)

func GrafanaAgentExtraSecretName() string {
	return grafanaAgentExtraSecretName
}

func IsLoggingEnabled(cluster *capi.Cluster, enableLoggingFlag bool) bool {
	// Logging should be enabled when all conditions are met:
	//   - logging label is set and true on the cluster
	//   - cluster is not being deleted
	//   - global logging flag is enabled

	// If the cluster is being deleted, always return false to trigger reconcileDelete
	// The delete logic will clean up everything regardless of previous logging state
	if !cluster.GetDeletionTimestamp().IsZero() {
		return false
	}

	// If logging is disabled at the installation level, return false
	if !enableLoggingFlag {
		return false
	}

	// Check cluster-specific logging label
	labels := cluster.GetLabels()
	loggingLabelValue, ok := labels[key.LoggingLabel]
	if !ok {
		return LoggingEnabledDefault
	}

	loggingEnabled, err := strconv.ParseBool(loggingLabelValue)
	if err != nil {
		return LoggingEnabledDefault
	}
	return loggingEnabled
}

func AddCommonLabels(labels map[string]string) {
	labels["giantswarm.io/managed-by"] = "logging-operator"
}

func IsWorkloadCluster(installationName, clusterName string) bool {
	return installationName != clusterName
}

// AppConfigName generates an app config name for the given cluster and app.
// This function can work with any cluster object.
func AppConfigName(cluster *capi.Cluster, app string) string {
	return fmt.Sprintf("%s-%s", cluster.GetName(), app)
}

// Read Loki URL from ingress
func ReadLokiIngressURL(ctx context.Context, cluster *capi.Cluster, client client.Client) (string, error) {
	var lokiIngress netv1.Ingress

	var objectKey = types.NamespacedName{Name: lokiGatewayIngressName, Namespace: lokiGatewayIngressNamespace}
	if err := client.Get(ctx, objectKey, &lokiIngress); err != nil {
		return "", errors.WithStack(err)
	}

	// We consider there's only one rule with one URL, because that's how the helm chart does it for the moment.
	if len(lokiIngress.Spec.Rules) <= 0 {
		return "", fmt.Errorf("loki ingress host not found")
	}
	return lokiIngress.Spec.Rules[0].Host, nil
}

// Read Tempo URL from ingress
func ReadTempoIngressURL(ctx context.Context, cluster *capi.Cluster, client client.Client) (string, error) {
	var tempoIngress netv1.Ingress

	var objectKey = types.NamespacedName{Name: tempoIngressName, Namespace: tempoIngressNamespace}
	if err := client.Get(ctx, objectKey, &tempoIngress); err != nil {
		return "", errors.WithStack(err)
	}

	// We consider there's only one rule with one URL, because that's how the helm chart does it for the moment.
	if len(tempoIngress.Spec.Rules) <= 0 {
		return "", fmt.Errorf("tempo ingress host not found")
	}
	return tempoIngress.Spec.Rules[0].Host, nil
}
