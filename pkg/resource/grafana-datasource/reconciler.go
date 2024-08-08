package grafanadatasource

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Reconciler implements a reconciler.Interface to handle
// Grafana datasource secret: secret which stores Loki datasource information for Grafana
type Reconciler struct {
	client.Client
}

// ReconcileCreate ensures Grafana Datasource for Loki is created with the right credentials
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("grafanadatasource create")

	// Retrieve secret containing credentials
	var loggingCredentialsSecret v1.Secret
	err := r.Client.Get(ctx, types.NamespacedName{Name: loggingcredentials.LoggingCredentialsSecretMeta(lc).Name, Namespace: loggingcredentials.LoggingCredentialsSecretMeta(lc).Namespace},
		&loggingCredentialsSecret)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Get desired secret
	desiredDatasourceSecret, err := GenerateDatasourceSecret(lc, &loggingCredentialsSecret)
	if err != nil {
		logger.Info("grafanadatasource - failed generating Datasource!", "error", err)
		return ctrl.Result{}, errors.WithStack(err)
	}

	err = common.EnsureCreatedOrUpdated(ctx, r.Client, desiredDatasourceSecret, needUpdate, "grafanadatasource")
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info("grafanadatasource - done")
	return ctrl.Result{}, nil
}

// ReconcileDelete - Not much to do here when a cluster is deleted
func (r *Reconciler) ReconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("grafanadatasource delete")

	return ctrl.Result{}, nil
}

// needUpdate return true if current.Data and desired.Data do not match.
func needUpdate(current, desired v1.Secret) bool {
	return !reflect.DeepEqual(current.Data, desired.Data)
}
