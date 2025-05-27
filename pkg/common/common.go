package common

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

const (
	// ReadUser is the global user for reading logs
	ReadUser = "read"
	// DefaultWriteTenant is the default tenant for writing logs
	DefaultWriteTenant = "giantswarm"
	// Loki Gateway Ingress
	lokiGatewayIngressNamespace = "loki"
	lokiGatewayIngressName      = "loki-gateway"
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

func IsLoggingEnabled(lc loggedcluster.Interface) bool {
	// Logging should be enabled when all conditions are met:
	//   - logging label is set and true on the cluster
	//   - cluster is not being deleted
	//   - global logging flag is enabled
	return lc.HasLoggingEnabled() && lc.GetDeletionTimestamp().IsZero() && lc.GetEnableLoggingFlag()
}

func AddCommonLabels(labels map[string]string) {
	labels["giantswarm.io/managed-by"] = "logging-operator"
}

func IsWorkloadCluster(lc loggedcluster.Interface) bool {
	return lc.GetInstallationName() != lc.GetClusterName()
}

// Read Loki URL from ingress
func ReadLokiIngressURL(ctx context.Context, lc loggedcluster.Interface, client client.Client) (string, error) {
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
