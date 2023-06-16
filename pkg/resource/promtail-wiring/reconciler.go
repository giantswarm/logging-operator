package promtailwiring

import (
	"context"
	"fmt"

	appv1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	promtailtoggle "github.com/giantswarm/logging-operator/pkg/resource/promtail-toggle"
	"github.com/pkg/errors"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Reconciler implements a reconciler.Interface to handle
// Promtail wiring: set or unset the user value configmap created by
// promtail-toggle in the observability bundle.
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// ReconcileCreate ensure user value configmap is set in observability bundle
// for the given cluster.
func (r *Reconciler) ReconcileCreate(ctx context.Context, object client.Object) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("promtailwiring create")

	// Get observability bundle app metadata.
	appMeta := ObservabilityBundleAppMeta(object)

	// Retrieve the app.
	logger.Info(fmt.Sprintf("promtailwiring checking %s/%s", appMeta.GetNamespace(), appMeta.GetNamespace()))
	var currentApp appv1.App
	err := r.Client.Get(ctx, types.NamespacedName{Name: appMeta.GetName(), Namespace: appMeta.GetNamespace()}, &currentApp)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Set user value configmap in the app.
	if setUserConfig(&currentApp, object) {
		logger.Info("promtailwiring updating")
		// Update the app.
		err := r.Client.Update(ctx, &currentApp)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	} else {
		logger.Info("promtailwiring up to date")
	}

	return ctrl.Result{}, nil
}

// ReconcileCreate ensure user value configmap is unset in observability bundle
// for the given cluster.
func (r *Reconciler) ReconcileDelete(ctx context.Context, object client.Object) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("promtailwiring delete")

	// Get observability bundle app metadata.
	appMeta := ObservabilityBundleAppMeta(object)

	var currentApp appv1.App
	err := r.Client.Get(ctx, types.NamespacedName{Name: appMeta.GetName(), Namespace: appMeta.GetNamespace()}, &currentApp)
	if err != nil {
		// Handle case where the app is not found.
		if apimachineryerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Unset user value configmap in the app.
	if unsetUserConfig(&currentApp, object) {
		logger.Info("promtailwiring updating")
		// Update the app.
		err = r.Client.Update(ctx, &currentApp)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	} else {
		logger.Info("promtailwiring up to date")
	}

	return ctrl.Result{}, nil
}

// setUserConfig set the user value confimap in the app.
// It returns true in case something was changed.
func setUserConfig(app *appv1.App, object client.Object) bool {
	observabilityBundleConfigMapMeta := promtailtoggle.ObservabilityBundleConfigMapMeta(object)
	updated := app.Spec.UserConfig.ConfigMap.Name != observabilityBundleConfigMapMeta.GetName() || app.Spec.UserConfig.ConfigMap.Namespace != observabilityBundleConfigMapMeta.GetNamespace()

	app.Spec.UserConfig.ConfigMap.Name = observabilityBundleConfigMapMeta.GetName()
	app.Spec.UserConfig.ConfigMap.Namespace = observabilityBundleConfigMapMeta.GetNamespace()

	return updated
}

// unsetUserConfig unset the user value confimap in the app.
// It returns true in case something was changed.
func unsetUserConfig(app *appv1.App, object client.Object) bool {
	observabilityBundleConfigMapMeta := promtailtoggle.ObservabilityBundleConfigMapMeta(object)
	updated := app.Spec.UserConfig.ConfigMap.Name == observabilityBundleConfigMapMeta.GetName() || app.Spec.UserConfig.ConfigMap.Namespace == observabilityBundleConfigMapMeta.GetNamespace()

	app.Spec.UserConfig.ConfigMap.Name = ""
	app.Spec.UserConfig.ConfigMap.Namespace = ""

	return updated
}
