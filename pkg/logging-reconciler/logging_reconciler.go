package loggingreconciler

import (
	"context"
	"fmt"

	"github.com/giantswarm/logging-operator/pkg/common"
	"github.com/giantswarm/logging-operator/pkg/key"
	"github.com/giantswarm/logging-operator/pkg/reconciler"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// LoggingReconciler reconciles logging for any supported object
type LoggingReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	Reconcilers []reconciler.Interface
}

func (l *LoggingReconciler) Reconcile(ctx context.Context, object client.Object) (result ctrl.Result, err error) {

	// Logging should be disable in case:
	//   - logging is disabled via a label on the Cluster object
	//   - Cluster object is being deleted
	loggingEnabled := clusterhelpers.IsLoggingEnabled(object) && object.GetDeletionTimestamp().IsZero()

	if loggingEnabled {

		// TODO: manage result
		_, err = l.reconcileCreate(ctx, object)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	} else {

		_, err = l.reconcileDelete(ctx, object)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	return ctrl.Result{}, nil
}

// reconcileCreate handles creation/update logic by calling ReconcileCreate method on all l.Reconcilers.
func (l *LoggingReconciler) reconcileCreate(ctx context.Context, object client.Object) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("LOGGING enabled")

	// Finalizer handling needs to come first.
	logger.Info(fmt.Sprintf("checking finalizer %s", key.Finalizer))
	if !controllerutil.ContainsFinalizer(object, key.Finalizer) {
		logger.Info(fmt.Sprintf("adding finalizer %s", key.Finalizer))
		controllerutil.AddFinalizer(object, key.Finalizer)
		err := l.Client.Update(ctx, object)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	// Call all reconcilers ReconcileCreate methods.
	for _, reconciler := range l.Reconcilers {
		// TODO(theo): add handling for returned ctrl.Result value.
		_, err := reconciler.ReconcileCreate(ctx, object)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	return ctrl.Result{}, nil
}

// reconcileDelete handles deletion logic by calling reconcileDelete method on all l.Reconcilers.
func (l *LoggingReconciler) reconcileDelete(ctx context.Context, object client.Object) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("LOGGING disabled")

	// Call all reconcilers ReconcileDelete methods.
	for _, reconciler := range l.Reconcilers {
		// TODO(theo): add handling for returned ctrl.Result value.
		_, err := reconciler.ReconcileDelete(ctx, object)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	// Finalizer handling needs to come last.
	logger.Info(fmt.Sprintf("checking finalizer %s", key.Finalizer))
	if controllerutil.ContainsFinalizer(object, key.Finalizer) {
		logger.Info(fmt.Sprintf("removing finalizer %s", key.Finalizer))
		controllerutil.RemoveFinalizer(object, key.Finalizer)
		err := l.Client.Update(ctx, object)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	return ctrl.Result{}, nil
}
