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
	Client client.Client
}

// ReconcileCreate ensures grafana-multi-tenant-proxy auth map is created with the right credentials
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	// If we are on CAPI, we don't need to create the proxyauth secret as we are not using the multi-tenant-proxy
	if lc.IsCAPI() {
		return r.ReconcileDelete(ctx, lc)
	}

	logger := log.FromContext(ctx)
	logger.Info("creating multi-tenant-proxy auth secret")

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
		logger.Error(err, "failed to generate multi-tenant-proxy auth secret")
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Check if auth config already exists.
	var currentProxyAuthSecret v1.Secret
	err = r.Client.Get(ctx, types.NamespacedName{Name: desiredProxyAuthSecret.GetName(), Namespace: desiredProxyAuthSecret.GetNamespace()}, &currentProxyAuthSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("multi-tenant-proxy auth secret not found, creating")
			err = r.Client.Create(ctx, &desiredProxyAuthSecret)
			if err != nil {
				return ctrl.Result{}, errors.WithStack(err)
			}
		} else {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	if !needUpdate(currentProxyAuthSecret, desiredProxyAuthSecret) {
		logger.Info("multi-tenant-proxy auth secret is up to date")
		return ctrl.Result{}, nil
	}

	logger.Info("updating multi-tenant-proxy auth secret")
	err = r.Client.Update(ctx, &desiredProxyAuthSecret)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info("updated multi-tenant-proxy auth secret")
	return ctrl.Result{}, nil
}

// ReconcileDelete - Delete the multi tenant proxy secret
func (r *Reconciler) ReconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("deleting multi-tenant-proxy auth secret")

	secret := secret()
	// Delete secret.
	err := r.Client.Delete(ctx, &secret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			// Do no throw error in case it was not found, as this means
			// it was already deleted.
			logger.Info("multi-tenant-proxy auth secret already deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}
	logger.Info("deleted multi-tenant-proxy auth secret")

	return ctrl.Result{}, nil
}

// needUpdate return true if current.Data and desired.Data do not match.
func needUpdate(current, desired v1.Secret) bool {
	return !reflect.DeepEqual(current.Data, desired.Data)
}
