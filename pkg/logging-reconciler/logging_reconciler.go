package loggingreconciler

import (
	"context"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/giantswarm/logging-operator/pkg/common"
	"github.com/giantswarm/logging-operator/pkg/key"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	"github.com/giantswarm/logging-operator/pkg/reconciler"
)

// LoggingReconciler reconciles logging for any supported object
type LoggingReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	Reconcilers []reconciler.Interface
}

func (l *LoggingReconciler) Reconcile(ctx context.Context, lc loggedcluster.Interface) (result ctrl.Result, err error) {
	if common.IsLoggingEnabled(lc) {
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
		controllerutil.AddFinalizer(lc, key.Finalizer)
		err := l.Client.Update(ctx, lc.GetObject())
		if err != nil {
			logger.Error(err, "failed to add finalizer to logger cluster", "finalizer", key.Finalizer)
			return ctrl.Result{}, errors.WithStack(err)
		}
		logger.Info("successfully added finalizer to logged cluster", "finalizer", key.Finalizer)
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
		controllerutil.RemoveFinalizer(lc, key.Finalizer)
		err := l.Client.Update(ctx, lc.GetObject())
		if err != nil {
			// We need to requeue if we fail to remove the finalizer because of race conditions between multiple operators.
			// This will be eventually consistent.
			logger.Error(err, "failed to remove finalizer from logger cluster, requeuing", "finalizer", key.Finalizer)
			return ctrl.Result{Requeue: true}, nil
		}
		logger.Info("successfully removed finalizer from logged cluster", "finalizer", key.Finalizer)
	}

	return ctrl.Result{}, nil
}
