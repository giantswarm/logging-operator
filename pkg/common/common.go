package common

import (
	"context"

	appv1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/pkg/errors"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

const (
	// ReadUser is the global user for reading logs
	ReadUser = "read"
	// Loki Ingress
	lokiIngressNamespace = "loki"
	lokiIngressName      = "loki-gateway"
	// grafana-agent secret name
	//#nosec G101
	grafanaAgentExtraSecretName = "grafana-agent-secret"

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

type lokiIngressURLContextKey int

var lokiIngressURLKey lokiIngressURLContextKey

// ReadLokiIngressURL reads the Loki Ingress URL and caches it in the context.
func ReadLokiIngressURL(ctx context.Context, lc loggedcluster.Interface, client client.Client) (string, error) {
	ingressURL, ok := ctx.Value(lokiIngressURLKey).(string)
	if ok {
		return ingressURL, nil
	}

	var lokiIngress netv1.Ingress

	err := client.Get(ctx, types.NamespacedName{Name: lokiIngressName, Namespace: lokiIngressNamespace}, &lokiIngress)
	if err != nil {
		return "", errors.WithStack(err)
	}

	// We consider there's only one rule with one URL, because that's how the helm chart does it for the moment.
	ingressURL = lokiIngress.Spec.Rules[0].Host

	ctx = context.WithValue(ctx, lokiIngressURLKey, ingressURL)

	return ingressURL, nil
}

type grafanaAgentAppContextKey int

var grafanaAgentAppKey grafanaAgentAppContextKey

// ReadGrafanaAgentApp reads the Grafana Agent app and caches it in the context.
func ReadGrafanaAgentApp(ctx context.Context, lc loggedcluster.Interface, client client.Client) (appv1.App, error) {
	var currentApp appv1.App

	currentApp, ok := ctx.Value(grafanaAgentAppKey).(appv1.App)
	if ok {
		return currentApp, nil
	}

	appMeta := ObservabilityBundleAppMeta(lc)

	// Check existence of grafana-agent app
	err := client.Get(ctx, types.NamespacedName{Name: lc.AppConfigName("grafana-agent"), Namespace: appMeta.GetNamespace()}, &currentApp)
	if err != nil {
		return appv1.App{}, errors.WithStack(err)
	}

	ctx = context.WithValue(ctx, grafanaAgentAppKey, currentApp)

	return currentApp, nil
}
