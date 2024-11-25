package eventsloggersecret

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
// Events-logger secret: extra events-logger secret about where and how to send logs (in this case : k8S events)
type Reconciler struct {
	client.Client
}

// ReconcileCreate ensures events-logger-secret is created with the right credentials
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("events-logger-secret create")

	// Retrieve secret containing credentials
	var eventsLoggerCredentialsSecret v1.Secret
	err := r.Client.Get(ctx, types.NamespacedName{Name: loggingcredentials.LoggingCredentialsSecretMeta().Name, Namespace: loggingcredentials.LoggingCredentialsSecretMeta().Namespace},
		&eventsLoggerCredentialsSecret)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Retrieve Loki ingress name
	lokiURL, err := common.ReadProxyIngressURL(ctx, lc, r.Client)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Get desired secret
	desiredEventsLoggerSecret, err := generateEventsLoggerSecret(lc, &eventsLoggerCredentialsSecret, lokiURL)
	if err != nil {
		logger.Error(err, "failed generating events logger secret")
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Check if secret already exists.
	logger.Info("events-logger-secret - getting", "namespace", desiredEventsLoggerSecret.GetNamespace(), "name", desiredEventsLoggerSecret.GetName())
	var currentEventsLoggerSecret v1.Secret
	err = r.Client.Get(ctx, types.NamespacedName{Name: desiredEventsLoggerSecret.GetName(), Namespace: desiredEventsLoggerSecret.GetNamespace()}, &currentEventsLoggerSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("events-logger-secret not found, creating")
			err = r.Client.Create(ctx, &desiredEventsLoggerSecret)
			if err != nil {
				return ctrl.Result{}, errors.WithStack(err)
			}
		} else {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	if !needUpdate(currentEventsLoggerSecret, desiredEventsLoggerSecret) {
		logger.Info("events-logger-secret up to date")
		return ctrl.Result{}, nil
	}

	logger.Info("updating events-logger-secret")
	err = r.Client.Update(ctx, &desiredEventsLoggerSecret)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info("updated events-logger-secret")
	return ctrl.Result{}, nil
}

// ReconcileDelete - Not much to do here when a cluster is deleted
func (r *Reconciler) ReconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("events-logger-secret delete")

	// Get expected secret.
	var currentEventsLoggerSecret v1.Secret
	err := r.Client.Get(ctx, types.NamespacedName{Name: GetEventsLoggerSecretName(lc), Namespace: lc.GetAppsNamespace()}, &currentEventsLoggerSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("events-logger-secret not found, stop here")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Delete secret.
	logger.Info("events-logger-secret deleting", "namespace", currentEventsLoggerSecret.GetNamespace(), "name", currentEventsLoggerSecret.GetName())
	err = r.Client.Delete(ctx, &currentEventsLoggerSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			// Do no throw error in case it was not found, as this means
			// it was already deleted.
			logger.Info("events-logger-secret already deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}
	logger.Info("events-logger-secret deleted")

	return ctrl.Result{}, nil
}

// needUpdate return true if current.Data and desired.Data do not match.
func needUpdate(current, desired v1.Secret) bool {
	return !reflect.DeepEqual(current.Data, desired.Data)
}
