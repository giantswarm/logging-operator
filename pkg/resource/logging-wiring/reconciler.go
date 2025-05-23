package loggingwiring

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
// Logging wiring: set or unset the user value configmap created by
// logging-agents-toggle in the observability bundle.
type Reconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
}

// ReconcileCreate ensure user value configmap is set in observability bundle
// for the given cluster.
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("logging wiring create")

	// Get observability bundle app metadata.
	appMeta := common.ObservabilityBundleAppMeta(lc)
	// Retrieve the app.
	logger.Info("logging wiring checking app", "namespace", appMeta.GetNamespace(), "name", appMeta.GetName())
	var currentApp appv1.App
	err := r.Client.Get(ctx, types.NamespacedName{Name: appMeta.GetName(), Namespace: appMeta.GetNamespace()}, &currentApp)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("logging wiring - app not found, requeuing")
			// If the app is not found we should requeue and try again later (5 minutes is the app platform default reconciliation time)
			return ctrl.Result{RequeueAfter: time.Duration(5 * time.Minute)}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	desiredApp := lc.WireLogging(currentApp)
	if !reflect.DeepEqual(currentApp, *desiredApp) {
		logger.Info("logging wiring updating")
		// Update the app.
		err := r.Client.Update(ctx, desiredApp)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	logger.Info("logging wiring up to date")

	return ctrl.Result{}, nil
}

// ReconcileCreate ensure user value configmap is unset in observability bundle
// for the given cluster.
func (r *Reconciler) ReconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("logging wiring delete")

	// Get observability bundle app metadata.
	appMeta := common.ObservabilityBundleAppMeta(lc)
	var currentApp appv1.App
	err := r.Client.Get(ctx, types.NamespacedName{Name: appMeta.GetName(), Namespace: appMeta.GetNamespace()}, &currentApp)
	if err != nil {
		// Handle case where the app is not found.
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("logging wiring - app not found, skipping deletion")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	desiredApp := lc.UnwireLogging(currentApp)
	if !reflect.DeepEqual(currentApp, *desiredApp) {
		logger.Info("logging wiring updating")
		// Update the app.
		err := r.Client.Update(ctx, desiredApp)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	logger.Info("logging wiring up to date")

	return ctrl.Result{}, nil
}
