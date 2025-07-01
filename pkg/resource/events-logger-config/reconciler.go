package eventsloggerconfig

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/giantswarm/logging-operator/pkg/config"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

// Reconciler implements a reconciler.Interface to handle
// EventsLogger config: extra events-logger config defining what we want to retrieve.
type Reconciler struct {
	Client            client.Client
	Config            config.Config
	IncludeNamespaces []string
	ExcludeNamespaces []string
}

// ReconcileCreate ensures events-logger config is created with the right credentials
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("events-logger-config create")

	// Get desired config
	desiredEventsLoggerConfig, err := generateEventsLoggerConfig(lc, r.IncludeNamespaces, r.ExcludeNamespaces, r.Config.InstallationName, r.Config.InsecureCA)
	if err != nil {
		logger.Info("events-logger-config - failed generating events-logger config!", "error", err)
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Check if config already exists.
	logger.Info("events-logger-config - getting", "namespace", desiredEventsLoggerConfig.GetNamespace(), "name", desiredEventsLoggerConfig.GetName())
	var currentEventsLoggerConfig v1.ConfigMap
	err = r.Client.Get(ctx, types.NamespacedName{Name: desiredEventsLoggerConfig.GetName(), Namespace: desiredEventsLoggerConfig.GetNamespace()}, &currentEventsLoggerConfig)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("events-logger-config not found, creating")
			err = r.Client.Create(ctx, &desiredEventsLoggerConfig)
			if err != nil {
				return ctrl.Result{}, errors.WithStack(err)
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	if !needUpdate(currentEventsLoggerConfig, desiredEventsLoggerConfig) {
		logger.Info("events-logger-config up to date")
		return ctrl.Result{}, nil
	}

	logger.Info("events-logger-config - updating")
	err = r.Client.Update(ctx, &desiredEventsLoggerConfig)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info("events-logger-config - done")
	return ctrl.Result{}, nil

}

func (r *Reconciler) ReconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("events-logger-config delete")

	// Get expected configmap.
	var currentEventsLoggerConfig v1.ConfigMap
	err := r.Client.Get(ctx, types.NamespacedName{Name: getEventsLoggerConfigName(lc), Namespace: lc.GetAppsNamespace()}, &currentEventsLoggerConfig)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("events-logger-config not found, stop here")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Delete configmap.
	logger.Info("events-logger-config deleting", "namespace", currentEventsLoggerConfig.GetNamespace(), "name", currentEventsLoggerConfig.GetName())
	err = r.Client.Delete(ctx, &currentEventsLoggerConfig)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			// Do no throw error in case it was not found, as this means
			// it was already deleted.
			logger.Info("events-logger-config already deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}
	logger.Info("events-logger-config deleted")

	return ctrl.Result{}, nil
}

// needUpdate return true if current.Data and desired.Data do not match.
func needUpdate(current, desired v1.ConfigMap) bool {
	return !reflect.DeepEqual(current.Data, desired.Data)
}
