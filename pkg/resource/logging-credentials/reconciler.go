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
	logger.Info("loggingCredentials secret create")

	const secretNamespace = "monitoring"
	const secretName = "logging-credentials"

	// Check if secrets exist.
	logger.Info(fmt.Sprintf("loggingcredentials checking secret %s/%s", secretNamespace, secretName))

	var currentSecret v1.Secret

	// To be updated when secret is modified
	var secretUpdated bool = false

	err := r.Client.Get(ctx, types.NamespacedName{Name: secretName, Namespace: secretNamespace}, &currentSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("logging-credentials not found, initializing one")
			// Create basic secret
			currentSecret, err = GenerateLoggingCredentialsBasicSecret(lc)
			secretUpdated = true
		}
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	// Check we have required data
	if _, ok := currentSecret.Data["readuser"]; !ok {
		logger.Info("loggingCredentials - adding readuser key")
		currentSecret.Data["readuser"] = []byte("read")
		secretUpdated = true
	}
	if _, ok := currentSecret.Data["readpassword"]; !ok {
		logger.Info("loggingCredentials - Adding readpassword key")
		currentSecret.Data["readpassword"] = []byte(genPassword())
		secretUpdated = true
	}
	if _, ok := currentSecret.Data["writeuser"]; !ok {
		logger.Info("loggingCredentials - Adding writeuser key")
		currentSecret.Data["writeuser"] = []byte("write")
		secretUpdated = true
	}
	if _, ok := currentSecret.Data["writepassword"]; !ok {
		logger.Info("loggingCredentials - Adding writepassword")
		currentSecret.Data["writepassword"] = []byte(genPassword())
		secretUpdated = true
	}

	if secretUpdated {
		logger.Info("loggingCredentials - Updating secret")
		err = r.Client.Update(ctx, &currentSecret)
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("loggingCredentials - Secret does not exist, creating it")
			err = r.Client.Create(ctx, &currentSecret)
		}
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
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
