package loggingcredentials

import (
	"context"
	"fmt"

	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	"github.com/pkg/errors"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Reconciler implements a reconciler.Interface to handle
// Logging Credentials: store and maintain logging credentials
type Reconciler struct {
	client.Client
}

// ReconcileCreate ensures a secret exists for the given cluster.
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info(fmt.Sprintf("loggingcredentials checking secret %s/%s", LoggingCredentialsSecretMeta(lc).Namespace, LoggingCredentialsSecretMeta(lc).Name))

	// Start with some empty secret
	loggingCredentialsSecret := GenerateLoggingCredentialsBasicSecret(lc)

	// Retrieve existing secret if it exists
	err := r.Client.Get(ctx, types.NamespacedName{Name: LoggingCredentialsSecretMeta(lc).Name, Namespace: LoggingCredentialsSecretMeta(lc).Namespace}, loggingCredentialsSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("loggingcredentials secret not found, initializing one")
		} else {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	// update the secret's contents if needed
	secretUpdated := UpdateLoggingCredentials(loggingCredentialsSecret)

	// commit our changes
	if secretUpdated {
		logger.Info("loggingCredentials - Updating secret")
		err = r.Client.Update(ctx, loggingCredentialsSecret)
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("loggingCredentials - Secret does not exist, creating it")
			err = r.Client.Create(ctx, loggingCredentialsSecret)
		}
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	} else {
		logger.Info("loggingCredentials - up to date")
	}

	return ctrl.Result{}, nil
}

// ReconcileDelete ensures a secret is removed for the current cluster
func (r *Reconciler) ReconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("loggingCredentials secret delete")

	// Well, for the moment we don't have per-cluster creds, so we won't delete any.

	return ctrl.Result{}, nil
}
