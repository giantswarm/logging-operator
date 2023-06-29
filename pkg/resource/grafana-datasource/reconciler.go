package grafanadatasource

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Reconciler implements a reconciler.Interface to handle
// Promtail toggle: enable or disable Promtail in a given Cluster.
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

	// Check if datasource already exists.
	logger.Info("grafanadatasource - getting", "namespace", desiredDatasourceSecret.GetNamespace(), "name", desiredDatasourceSecret.GetName())
	var currentDatasourceSecret v1.Secret
	err = r.Client.Get(ctx, types.NamespacedName{Name: desiredDatasourceSecret.GetName(), Namespace: desiredDatasourceSecret.GetNamespace()}, &currentDatasourceSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("grafanadatasource not found, creating")
			err = r.Client.Create(ctx, &desiredDatasourceSecret)
			if err != nil {
				return ctrl.Result{}, errors.WithStack(err)
			}
		} else {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	if !needUpdate(currentDatasourceSecret, desiredDatasourceSecret) {
		logger.Info("grafanadatasource up to date")
		return ctrl.Result{}, nil
	}

	logger.Info("grafanadatasource - updating")
	err = r.Client.Update(ctx, &desiredDatasourceSecret)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

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
