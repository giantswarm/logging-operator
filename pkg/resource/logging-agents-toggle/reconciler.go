package loggingagentstoggle

import (
	"context"
	"reflect"

	"github.com/blang/semver"
	appv1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Reconciler implements a reconciler.Interface to handle
// Logging agents toggle: enable or disable logging agents in a given Cluster.
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *Reconciler) GetObservabilityBundleVersion(ctx context.Context, lc loggedcluster.Interface) (semver.Version, error) {
	// Get observability bundle app metadata.
	appMeta := common.ObservabilityBundleAppMeta(lc)

	// Retrieve the app.
	var currentApp appv1.App
	err := r.Client.Get(ctx, types.NamespacedName{Name: appMeta.GetName(), Namespace: appMeta.GetNamespace()}, &currentApp)
	if err != nil {
		return semver.Version{}, errors.WithStack(err)
	}
	return semver.Parse(currentApp.Spec.Version)
}

// ReconcileCreate ensure logging agents are enabled in the given cluster.
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Logging agents toggle create")

	observabilityBundleVersion, err := r.GetObservabilityBundleVersion(ctx, lc)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Get desired configmap to enable logging agents.
	desiredConfigMap, err := GenerateObservabilityBundleConfigMap(lc, observabilityBundleVersion)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Check if configmap is already installed.
	logger.Info("Logging agents toggle checking", "namespace", desiredConfigMap.GetNamespace(), "name", desiredConfigMap.GetName())
	var currentConfigMap v1.ConfigMap
	err = r.Client.Get(ctx, types.NamespacedName{Name: desiredConfigMap.GetName(), Namespace: desiredConfigMap.GetNamespace()}, &currentConfigMap)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			// Install configmap.
			// Configmap was not found.
			logger.Info("Logging agents toggle not found, creating")
			err = r.Client.Create(ctx, &desiredConfigMap)
		}
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	if needUpdate(currentConfigMap, desiredConfigMap) {
		logger.Info("Logging agents toggle updating")
		// Update configmap
		// Configmap is installed and need to be updated.
		err := r.Client.Update(ctx, &desiredConfigMap)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	} else {
		logger.Info("Logging agents toggle up to date")
	}

	return ctrl.Result{}, nil
}

// ReconcileDelete ensure logging agents are disabled for the given cluster.
func (r *Reconciler) ReconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Logging agents toggle delete")

	observabilityBundleVersion, err := r.GetObservabilityBundleVersion(ctx, lc)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Get expected configmap.
	desiredConfigMap, err := GenerateObservabilityBundleConfigMap(lc, observabilityBundleVersion)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Delete configmap.
	logger.Info("Logging agents toggle deleting", "namespace", desiredConfigMap.GetNamespace(), "name", desiredConfigMap.GetName())
	err = r.Client.Delete(ctx, &desiredConfigMap)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			// Do no throw error in case it was not found, as this means
			// it was already deleted.
			logger.Info("Logging agents toggle already deleted")
		} else if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	} else {
		logger.Info("Logging agents toggle deleted")
	}

	return ctrl.Result{}, nil
}

// needUpdate return true if current.Data and desired.Data do not match.
func needUpdate(current, desired v1.ConfigMap) bool {
	return !reflect.DeepEqual(current.Data, desired.Data) || !reflect.DeepEqual(current.ObjectMeta.Labels, desired.ObjectMeta.Labels)
}
