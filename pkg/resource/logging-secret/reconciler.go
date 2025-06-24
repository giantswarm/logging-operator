package loggingsecret

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/logging-operator/pkg/common"
	"github.com/giantswarm/logging-operator/pkg/config"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Reconciler implements a reconciler.Interface to handle
// Logging secret: extra logging secret about where and how to send logs
type Reconciler struct {
	Client client.Client
	Config config.Config
}

// ReconcileCreate ensures logging-secret is created with the right credentials
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("logging-secret create")

	// Retrieve secret containing credentials
	var loggingCredentialsSecret v1.Secret
	err := r.Client.Get(ctx, types.NamespacedName{Name: loggingcredentials.LoggingCredentialsSecretMeta().Name, Namespace: loggingcredentials.LoggingCredentialsSecretMeta().Namespace},
		&loggingCredentialsSecret)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Retrieve Loki ingress name
	lokiURL, err := common.ReadLokiIngressURL(ctx, lc, r.Client)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Get desired secret
	desiredLoggingSecret, err := GenerateLoggingSecret(lc, &loggingCredentialsSecret, lokiURL, r.Config.InstallationName, r.Config.InsecureCA)
	if err != nil {
		logger.Info("logging-secret - failed generating auth config!", "error", err)
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Check if secret already exists.
	logger.Info("logging-secret - getting", "namespace", desiredLoggingSecret.GetNamespace(), "name", desiredLoggingSecret.GetName())
	var currentLoggingSecret v1.Secret
	err = r.Client.Get(ctx, types.NamespacedName{Name: desiredLoggingSecret.Name, Namespace: desiredLoggingSecret.Namespace}, &currentLoggingSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("logging-secret not found, creating")
			err = r.Client.Create(ctx, &desiredLoggingSecret)
			if err != nil {
				return ctrl.Result{}, errors.WithStack(err)
			}
		} else {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	if !needUpdate(currentLoggingSecret, desiredLoggingSecret) {
		logger.Info("logging-secret up to date")
		return ctrl.Result{}, nil
	}

	logger.Info("logging-secret - updating")
	err = r.Client.Update(ctx, &desiredLoggingSecret)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info("logging-secret - done")
	return ctrl.Result{}, nil
}

// ReconcileDelete - Not much to do here when a cluster is deleted
func (r *Reconciler) ReconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("logging-secret delete")

	// Get expected secret.
	var currentLoggingSecret v1.Secret
	err := r.Client.Get(ctx, types.NamespacedName{Name: GetLoggingSecretName(lc), Namespace: lc.GetNamespace()}, &currentLoggingSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("logging-secret not found, stop here")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Delete secret.
	logger.Info("logging-secret deleting", "namespace", currentLoggingSecret.GetNamespace(), "name", currentLoggingSecret.GetName())
	err = r.Client.Delete(ctx, &currentLoggingSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			// Do no throw error in case it was not found, as this means
			// it was already deleted.
			logger.Info("logging-secret already deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}
	logger.Info("logging-secret deleted")

	return ctrl.Result{}, nil
}

// needUpdate return true if current.Data and desired.Data do not match.
func needUpdate(current, desired v1.Secret) bool {
	return !reflect.DeepEqual(current.Data, desired.Data)
}
