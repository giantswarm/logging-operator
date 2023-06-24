package loggingcredentials

import (
	"context"
	"fmt"

	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Reconciler implements a reconciler.Interface to handle
// Logging Credentials: store and maintain logging credentials
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// ReconcileCreate ensures a secret exists for the given cluster.
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	const secretNamespace = "monitoring"
	const secretName = "logging-credentials"

	logger.Info(fmt.Sprintf("loggingcredentials checking secret %s/%s", secretNamespace, secretName))

	var loggingCredentialsSecret v1.Secret

	// Check if secrets exist / retrieve existing secret
	err := r.Client.Get(ctx, types.NamespacedName{Name: secretName, Namespace: secretNamespace}, &loggingCredentialsSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("logging-credentials not found, initializing one")
			// Create basic secret
			loggingCredentialsSecret, err = GenerateLoggingCredentialsBasicSecret(lc)
		}
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	// update the secret's contents if needed
	secretUpdated := UpdateLoggingCredentials(loggingCredentialsSecret)

	// commit our changes
	if secretUpdated {
		logger.Info("loggingCredentials - Updating secret")
		err = r.Client.Update(ctx, &loggingCredentialsSecret)
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("loggingCredentials - Secret does not exist, creating it")
			err = r.Client.Create(ctx, &loggingCredentialsSecret)
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
