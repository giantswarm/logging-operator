package common

import (
	"context"
	"time"

	"github.com/pkg/errors"
	netv1 "k8s.io/api/networking/v1"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appv1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"

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

// NewLokiIngressURLContext reads the Loki Ingress URL from the API and stores it in the context.
func NewLokiIngressURLContext(ctx context.Context, lc loggedcluster.Interface, client client.Client) (context.Context, ctrl.Result, error) {
	var lokiIngress netv1.Ingress

	err := client.Get(ctx, types.NamespacedName{Name: lokiIngressName, Namespace: lokiIngressNamespace}, &lokiIngress)
	if err != nil {
		return nil, ctrl.Result{}, errors.WithStack(err)
	}

	// We consider there's only one rule with one URL, because that's how the helm chart does it for the moment.
	ingressURL := lokiIngress.Spec.Rules[0].Host

	ctx = context.WithValue(ctx, lokiIngressURLKey, ingressURL)

	return ctx, ctrl.Result{}, nil
}

func LokiIngressURLFromContext(ctx context.Context) (string, bool) {
	lokiIngressURL, ok := ctx.Value(lokiIngressURLKey).(string)
	return lokiIngressURL, ok
}

type grafanaAgentAppContextKey int

var grafanaAgentAppKey grafanaAgentAppContextKey

// NewGrafanaAgentAppContext retrieves the grafana-agent app from the API and stores it in the context.
func NewGrafanaAgentAppContext(ctx context.Context, lc loggedcluster.Interface, client client.Client) (context.Context, ctrl.Result, error) {
	var currentApp appv1.App

	appMeta := ObservabilityBundleAppMeta(lc)

	// Check existence of grafana-agent app
	err := client.Get(ctx, types.NamespacedName{Name: lc.AppConfigName("grafana-agent"), Namespace: appMeta.GetNamespace()}, &currentApp)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			// If the app is not found we should requeue and try again later (5 minutes is the app platform default reconciliation time)
			return nil, ctrl.Result{RequeueAfter: time.Duration(5 * time.Minute)}, nil
		}
		return nil, ctrl.Result{}, errors.WithStack(err)
	}

	ctx = context.WithValue(ctx, grafanaAgentAppKey, currentApp)

	return ctx, ctrl.Result{}, nil
}

func GrafanaAgentAppFromContext(ctx context.Context) (appv1.App, bool) {
	grafanaAgentApp, ok := ctx.Value(grafanaAgentAppKey).(appv1.App)
	return grafanaAgentApp, ok
}
