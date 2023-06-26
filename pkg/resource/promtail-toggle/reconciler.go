package promtailtoggle

import (
	"context"
	"fmt"
	"reflect"

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
// Promtail toggle: enable or disable Promtail in a given Cluster.
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// ReconcileCreate ensure Promtail is enabled in the given cluster.
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("promtailtoggle create")

	// Get desired configmap to enable promtail.
	desiredConfigMap, err := GenerateObservabilityBundleConfigMap(lc)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Check if configmap is already installed.
	logger.Info(fmt.Sprintf("promtailtoggle checking %s/%s", desiredConfigMap.GetNamespace(), desiredConfigMap.GetName()))
	var currentConfigMap v1.ConfigMap
	err = r.Client.Get(ctx, types.NamespacedName{Name: desiredConfigMap.GetName(), Namespace: desiredConfigMap.GetNamespace()}, &currentConfigMap)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			// Install configmap.
			// Configmap was not found.
			logger.Info("promtailtoggle not found, creating")
			err = r.Client.Create(ctx, &desiredConfigMap)
		}
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	if needUpdate(currentConfigMap, desiredConfigMap) {
		logger.Info("promtailtoggle updating")
		// Update configmap
		// Configmap is installed and need to be updated.
		err := r.Client.Update(ctx, &desiredConfigMap)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	} else {
		logger.Info("promtailtoggle up to date")
	}

	return ctrl.Result{}, nil
}

// ReconcileDelete ensure Promtail is disabled for the given cluster.
func (r *Reconciler) ReconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("promtailtoggle delete")

	// Get expected configmap.
	desiredConfigMap, err := GenerateObservabilityBundleConfigMap(lc)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Delete configmap.
	logger.Info(fmt.Sprintf("promtailtoggle deleting %s/%s", desiredConfigMap.GetNamespace(), desiredConfigMap.GetName()))
	err = r.Client.Delete(ctx, &desiredConfigMap)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			// Do no throw error in case it was not found, as this means
			// it was already deleted.
			logger.Info("promtailtoggle already deleted")
		} else if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	} else {
		logger.Info("promtailtoggle deleted")
	}

	return ctrl.Result{}, nil
}

// needUpdate return true if current.Data and desired.Data do not match.
func needUpdate(current, desired v1.ConfigMap) bool {
	return !reflect.DeepEqual(current.Data, desired.Data) || !reflect.DeepEqual(current.ObjectMeta.Labels, desired.ObjectMeta.Labels)
}
