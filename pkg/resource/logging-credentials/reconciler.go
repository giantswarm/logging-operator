package loggingcredentials

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"

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

	logger.Info("loggingcredentials checking secret", "namespace", LoggingCredentialsSecretMeta(lc).Namespace, "name", LoggingCredentialsSecretMeta(lc).Name)

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
	secretUpdated, err := AddLoggingCredentials(lc, loggingCredentialsSecret)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Check if metadata has been updated
	if !reflect.DeepEqual(loggingCredentialsSecret.ObjectMeta.Labels, LoggingCredentialsSecretMeta(lc).Labels) {
		logger.Info("loggingCredentials - metatada update required")
		loggingCredentialsSecret.ObjectMeta = LoggingCredentialsSecretMeta(lc)
		secretUpdated = true
	}

	if !secretUpdated {
		// If there were no changes, we're done here.
		logger.Info("loggingCredentials - up to date")
		return ctrl.Result{}, nil
	}

	// commit our changes
	logger.Info("loggingCredentials - Updating secret")
	err = r.Client.Update(ctx, loggingCredentialsSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("loggingCredentials - Secret does not exist, creating it")
			err = r.Client.Create(ctx, loggingCredentialsSecret)
		}
	}

	// Will return Secret's update error if any
	return ctrl.Result{}, errors.WithStack(err)
}

// ReconcileDelete ensures a secret is removed for the current cluster
func (r *Reconciler) ReconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("loggingcredentials secret delete", "namespace", LoggingCredentialsSecretMeta(lc).Namespace, "name", LoggingCredentialsSecretMeta(lc).Name)

	// Start with some empty secret
	loggingCredentialsSecret := GenerateLoggingCredentialsBasicSecret(lc)

	// Retrieve existing secret
	err := r.Client.Get(ctx, types.NamespacedName{Name: LoggingCredentialsSecretMeta(lc).Name, Namespace: LoggingCredentialsSecretMeta(lc).Namespace}, loggingCredentialsSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("loggingcredentials secret not found, initializing one")
			return ctrl.Result{}, nil
		} else {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	// update the secret's contents if needed
	secretUpdated := RemoveLoggingCredentials(lc, loggingCredentialsSecret)

	// Check if metadata has been updated
	if !reflect.DeepEqual(loggingCredentialsSecret.ObjectMeta.Labels, LoggingCredentialsSecretMeta(lc).Labels) {
		logger.Info("loggingCredentials - metatada update required")
		loggingCredentialsSecret.ObjectMeta = LoggingCredentialsSecretMeta(lc)
		secretUpdated = true
	}

	if !secretUpdated {
		// If there were no changes, we're done here.
		logger.Info("loggingCredentials - up to date")
		return ctrl.Result{}, nil
	}

	// commit our changes
	logger.Info("loggingCredentials - Updating secret")
	err = r.Client.Update(ctx, loggingCredentialsSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("loggingCredentials - Secret does not exist, creating it")
			err = r.Client.Create(ctx, loggingCredentialsSecret)
		}
	}

	// Will return Secret's update error if any
	return ctrl.Result{}, errors.WithStack(err)
}
