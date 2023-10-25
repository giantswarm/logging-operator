package common

import (
	"context"

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
	grafanaAgentResourceName = "grafana-agent-secret"
)

func GetGrafanaAgentResourceName() string {
	return grafanaAgentResourceName
}

func IsLoggingEnabled(lc loggedcluster.Interface) bool {

	// Logging should be enabled when all conditions are met:
	//   - logging label is set and true on the cluster
	//   - cluster is not being deleted
	//   - TODO(theo) global logging flag is enabled

	return lc.GetLoggingLabel() == "true" && lc.GetDeletionTimestamp().IsZero() && lc.GetEnableLoggingFlag()
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

	err := client.Get(ctx, types.NamespacedName{Name: lokiIngressName, Namespace: lokiIngressNamespace}, &lokiIngress)
	if err != nil {
		return "", errors.WithStack(err)
	}

	// We consider there's only one rule with one URL, because that's how the helm chart does it for the moment.
	ingressURL := lokiIngress.Spec.Rules[0].Host

	return ingressURL, nil
}
