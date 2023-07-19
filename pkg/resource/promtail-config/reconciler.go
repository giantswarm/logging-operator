package promtailconfig

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Reconciler implements a reconciler.Interface to handle
// Promtail config: extra promtail config defining what we want to retrieve.
type Reconciler struct {
	client.Client
}

// ReconcileCreate ensures promtail config is created with the right credentials
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("promtailconfig create")

	// Get desired config
	desiredPromtailConfig, err := GeneratePromtailConfig(lc)
	if err != nil {
		logger.Info("promtailconfig - failed generating promtail config!", "error", err)
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Check if config already exists.
	logger.Info("promtailconfig - getting", "namespace", desiredPromtailConfig.GetNamespace(), "name", desiredPromtailConfig.GetName())
	var currentPromtailConfig v1.ConfigMap
	err = r.Client.Get(ctx, types.NamespacedName{Name: desiredPromtailConfig.GetName(), Namespace: desiredPromtailConfig.GetNamespace()}, &currentPromtailConfig)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("promtailconfig not found, creating")
			err = r.Client.Create(ctx, &desiredPromtailConfig)
			if err != nil {
				return ctrl.Result{}, errors.WithStack(err)
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	if !needUpdate(currentPromtailConfig, desiredPromtailConfig) {
		logger.Info("promtailconfig up to date")
		return ctrl.Result{}, nil
	}

	logger.Info("promtailconfig - updating")
	err = r.Client.Update(ctx, &desiredPromtailConfig)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info("promtailconfig - done")
	return ctrl.Result{}, nil
}

// ReconcileDelete - Not much to do here when a cluster is deleted
func (r *Reconciler) ReconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("promtailconfig delete")

	// Don't do anything, we let observability-bundle do the cleanup when logging.enable=false

	return ctrl.Result{}, nil
}

// needUpdate return true if current.Data and desired.Data do not match.
func needUpdate(current, desired v1.ConfigMap) bool {
	return !reflect.DeepEqual(current.Data, desired.Data)
}
