package logging

import (
	"context"
	"time"

	"github.com/pkg/errors"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/giantswarm/logging-operator/pkg/common"
	"github.com/giantswarm/logging-operator/pkg/config"
	"github.com/giantswarm/logging-operator/pkg/key"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	"github.com/giantswarm/logging-operator/pkg/reconciler"
)

// LoggingReconciler reconciles logging for any supported object
type LoggingReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	Reconcilers []reconciler.Interface
	Config      *config.Config
}

func (l *LoggingReconciler) Reconcile(ctx context.Context, lc loggedcluster.Interface) (result ctrl.Result, err error) {
	if common.IsLoggingEnabled(lc, l.Config.EnableLoggingFlag) {
		result, err = l.reconcileCreate(ctx, lc)
	} else {
		result, err = l.reconcileDelete(ctx, lc)
	}

	return result, errors.WithStack(err)
}

// reconcileCreate handles creation/update logic by calling ReconcileCreate method on all l.Reconcilers.
func (l *LoggingReconciler) reconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("LOGGING enabled")

	if !controllerutil.ContainsFinalizer(lc, key.Finalizer) {
		logger.Info("adding finalizer", "finalizer", key.Finalizer)

		// We use a patch rather than an update to avoid conflicts when multiple controllers are adding their finalizer to the ClusterCR
		// We use the patch from sigs.k8s.io/cluster-api/util/patch to handle the patching without conflicts
		patchHelper, err := patch.NewHelper(lc.GetObject(), l.Client)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
		controllerutil.AddFinalizer(lc, key.Finalizer)
		if err := patchHelper.Patch(ctx, lc.GetObject()); err != nil {
			logger.Error(err, "failed to add finalizer to logger cluster", "finalizer", key.Finalizer)
			return ctrl.Result{}, errors.WithStack(err)
		}
		logger.Info("successfully added finalizer to logged cluster", "finalizer", key.Finalizer)
	}

	err := common.ToggleAgents(ctx, l.Client, lc)
	if err != nil {
		// Handle case where the app is not found.
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("observability bundle app not found, requeueing")
			// If the app is not found we should requeue and try again later (5 minutes is the app platform default reconciliation time)
			return ctrl.Result{RequeueAfter: time.Duration(5 * time.Minute)}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Call all reconcilers ReconcileCreate methods.
	for _, reconciler := range l.Reconcilers {
		result, err := reconciler.ReconcileCreate(ctx, lc)
		if err != nil || !result.IsZero() {
			return result, errors.WithStack(err)
		}
	}

	return ctrl.Result{}, nil
}

// reconcileDelete handles deletion logic by calling reconcileDelete method on all l.Reconcilers.
func (l *LoggingReconciler) reconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("LOGGING disabled")

	if controllerutil.ContainsFinalizer(lc, key.Finalizer) {
		err := common.ToggleAgents(ctx, l.Client, lc)
		if err != nil && !apimachineryerrors.IsNotFound(err) {
			// Errors only if this is not a 404 because the apps are already deleted.
			return ctrl.Result{}, errors.WithStack(err)
		}

		// Call all reconcilers ReconcileDelete methods.
		for _, reconciler := range l.Reconcilers {
			result, err := reconciler.ReconcileDelete(ctx, lc)
			if err != nil || !result.IsZero() {
				return result, errors.WithStack(err)
			}
		}

		// We get the latest state of the object to avoid race conditions.
		// Finalizer handling needs to come last.
		logger.Info("removing finalizer", "finalizer", key.Finalizer)

		// We use a patch rather than an update to avoid conflicts when multiple controllers are removing their finalizer from the ClusterCR
		// We use the patch from sigs.k8s.io/cluster-api/util/patch to handle the patching without conflicts
		patchHelper, err := patch.NewHelper(lc.GetObject(), l.Client)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
		controllerutil.RemoveFinalizer(lc, key.Finalizer)
		if err := patchHelper.Patch(ctx, lc.GetObject()); err != nil {
			logger.Error(err, "failed to remove finalizer from logger cluster, requeuing", "finalizer", key.Finalizer)
			return ctrl.Result{}, errors.WithStack(err)
		}
		logger.Info("successfully removed finalizer from logged cluster", "finalizer", key.Finalizer)
	}

	return ctrl.Result{}, nil
}
