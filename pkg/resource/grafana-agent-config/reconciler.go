package grafanaagentconfig

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Reconciler implements a reconciler.Interface to handle
// GrafanaAgent config: extra grafana-agent config defining what we want to retrieve.
type Reconciler struct {
	client.Client
}

// ReconcileCreate ensures grafana-agent config is created with the right credentials
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("grafana-agent-config create")

	// Retrieve secret containing credentials
	var loggingCredentialsSecret v1.Secret
	err := r.Client.Get(ctx, types.NamespacedName{Name: loggingcredentials.LoggingCredentialsSecretMeta(lc).Name, Namespace: loggingcredentials.LoggingCredentialsSecretMeta(lc).Namespace},
		&loggingCredentialsSecret)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Retrieve Loki ingress name
	lokiURL, err := common.ReadLokiIngressURL(ctx, lc, r.Client)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Get desired config
	desiredGrafanaAgentConfig, err := GenerateGrafanaAgentConfig(lc, &loggingCredentialsSecret, lokiURL)
	if err != nil {
		logger.Info("grafana-agent-config - failed generating grafana-agent config!", "error", err)
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Check if config already exists.
	logger.Info("grafana-agent-config - getting", "namespace", desiredGrafanaAgentConfig.GetNamespace(), "name", desiredGrafanaAgentConfig.GetName())
	var currentGrafanaAgentConfig v1.ConfigMap
	err = r.Client.Get(ctx, types.NamespacedName{Name: desiredGrafanaAgentConfig.GetName(), Namespace: desiredGrafanaAgentConfig.GetNamespace()}, &currentGrafanaAgentConfig)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("grafana-agent-config not found, creating")
			err = r.Client.Create(ctx, &desiredGrafanaAgentConfig)
			if err != nil {
				return ctrl.Result{}, errors.WithStack(err)
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	if !needUpdate(currentGrafanaAgentConfig, desiredGrafanaAgentConfig) {
		logger.Info("grafana-agent-config up to date")
		return ctrl.Result{}, nil
	}

	logger.Info("grafana-agent-config - updating")
	err = r.Client.Update(ctx, &desiredGrafanaAgentConfig)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info("grafana-agent-config - done")
	return ctrl.Result{}, nil
}

// ReconcileDelete - Not much to do here when a cluster is deleted
func (r *Reconciler) ReconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("grafana-agent-config delete")

	// Don't do anything, we let observability-bundle do the cleanup when logging.enable=false

	return ctrl.Result{}, nil
}

// needUpdate return true if current.Data and desired.Data do not match.
func needUpdate(current, desired v1.ConfigMap) bool {
	return !reflect.DeepEqual(current.Data, desired.Data)
}
