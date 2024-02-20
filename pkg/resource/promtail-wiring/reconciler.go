package promtailwiring

import (
	"context"
	"reflect"
	"time"

	appv1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/pkg/errors"
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
// Promtail wiring: set or unset the user value configmap created by
// promtail-toggle in the observability bundle.
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// ReconcileCreate ensure user value configmap is set in observability bundle
// for the given cluster.
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("promtailwiring create")

	// Get observability bundle app metadata.
	appMeta := common.ObservabilityBundleAppMeta(lc)

	// Retrieve the app.
	logger.Info("promtailwiring checking app", "namespace", appMeta.GetNamespace(), "name", appMeta.GetName())
	var currentApp appv1.App
	err := r.Client.Get(ctx, types.NamespacedName{Name: appMeta.GetName(), Namespace: appMeta.GetNamespace()}, &currentApp)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			// If the app is not found we should requeue and try again later (5 minutes is the app platform default reconciliation time)
			return ctrl.Result{RequeueAfter: time.Duration(5 * time.Minute)}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	desiredApp := lc.WirePromtail(currentApp)
	if !reflect.DeepEqual(currentApp, *desiredApp) {
		logger.Info("promtailwiring updating")
		// Update the app.
		err := r.Client.Update(ctx, desiredApp)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	logger.Info("promtailwiring up to date")

	return ctrl.Result{}, nil
}

// ReconcileCreate ensure user value configmap is unset in observability bundle
// for the given cluster.
func (r *Reconciler) ReconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("promtailwiring delete")

	// Get observability bundle app metadata.
	appMeta := common.ObservabilityBundleAppMeta(lc)
	var currentApp appv1.App
	err := r.Client.Get(ctx, types.NamespacedName{Name: appMeta.GetName(), Namespace: appMeta.GetNamespace()}, &currentApp)
	if err != nil {
		// Handle case where the app is not found.
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("promtailwiring - app not found")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	desiredApp := lc.UnwirePromtail(currentApp)
	if !reflect.DeepEqual(currentApp, *desiredApp) {
		logger.Info("promtailwiring updating")
		// Update the app.
		err := r.Client.Update(ctx, desiredApp)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	logger.Info("promtailwiring up to date")

	return ctrl.Result{}, nil
}
