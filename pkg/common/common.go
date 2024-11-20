package common

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

const (
	// ReadUser is the global user for reading logs
	ReadUser = "read"
	// Grafana Multi Tenant Proxy Ingress
	proxyIngressNamespace = "monitoring"
	proxyIngressName      = "grafana-multi-tenant-proxy"
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
	AlloyObservabilityBundleAppName          = "alloyLogs"
	PromtailObservabilityBundleAppName       = "promtail"
	PromtailObservabilityBundleLegacyAppName = "promtail-app"

	// Alloy app name and namespace when using Alloy as logging agent.
	AlloyLogAgentAppName      = "alloy-logs"
	AlloyLogAgentAppNamespace = "kube-system"

	// Alloy app name and namespace when using ALloy as events logger.
	AlloyEventsLoggerAppName      = "alloy-events"
	AlloyEventsLoggerAppNamespace = "kube-system"

	MaxBackoffPeriod = "10m"
	LokiURLFormat    = "https://%s/loki/api/v1/push"
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

// Read Proxy URL from ingress
func ReadProxyIngressURL(ctx context.Context, lc loggedcluster.Interface, client client.Client) (string, error) {
	var proxyIngress netv1.Ingress

	err := client.Get(ctx, types.NamespacedName{Name: proxyIngressName, Namespace: proxyIngressNamespace}, &proxyIngress)
	if err != nil {
		return "", errors.WithStack(err)
	}

	// We consider there's only one rule with one URL, because that's how the helm chart does it for the moment.
	ingressURL := proxyIngress.Spec.Rules[0].Host

	return ingressURL, nil
}

func FormatScrapedNamespaces(lc loggedcluster.Interface, namespaces []string) string {
	if IsWorkloadCluster(lc) {
		return fmt.Sprintf("[\"%s\"]", strings.Join(namespaces, "\", \""))
	}

	return "[]"
}
