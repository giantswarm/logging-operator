package proxyauth

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
// Proxy auth: a secret for the grafana-multi-tenant-proxy config
type Reconciler struct {
	client.Client
}

// ReconcileCreate ensures grafana-multi-tenant-proxy auth map is created with the right credentials
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("proxyauth create")

	// Retrieve secret containing credentials
	var proxyAuthSecret v1.Secret
	err := r.Client.Get(ctx, types.NamespacedName{Name: loggingcredentials.LoggingCredentialsSecretMeta().Name, Namespace: loggingcredentials.LoggingCredentialsSecretMeta().Namespace},
		&proxyAuthSecret)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Get desired secret
	desiredProxyAuthSecret, err := GenerateProxyAuthSecret(lc, &proxyAuthSecret)
	if err != nil {
		logger.Info("proxyAuth - failed generating auth config!", "error", err)
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Check if auth config already exists.
	logger.Info("proxyAuth - getting", "namespace", desiredProxyAuthSecret.GetNamespace(), "name", desiredProxyAuthSecret.GetName())
	var currentProxyAuthSecret v1.Secret
	err = r.Client.Get(ctx, types.NamespacedName{Name: desiredProxyAuthSecret.GetName(), Namespace: desiredProxyAuthSecret.GetNamespace()}, &currentProxyAuthSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("proxyAuth not found, creating")
			err = r.Client.Create(ctx, &desiredProxyAuthSecret)
			if err != nil {
				return ctrl.Result{}, errors.WithStack(err)
			}
		} else {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	if !needUpdate(currentProxyAuthSecret, desiredProxyAuthSecret) {
		logger.Info("proxyauth up to date")
		return ctrl.Result{}, nil
	}

	logger.Info("proxyauth - updating")
	err = r.Client.Update(ctx, &desiredProxyAuthSecret)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info("proxyauth - done")
	return ctrl.Result{}, nil
}

// ReconcileDelete - Not much to do here when a cluster is deleted
func (r *Reconciler) ReconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("proxyAuth delete")

	return ctrl.Result{}, nil
}

// needUpdate return true if current.Data and desired.Data do not match.
func needUpdate(current, desired v1.Secret) bool {
	return !reflect.DeepEqual(current.Data, desired.Data)
}
