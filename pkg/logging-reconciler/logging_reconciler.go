package loggingreconciler

import (
	"context"
	"fmt"

	"github.com/giantswarm/logging-operator/pkg/common"
	"github.com/giantswarm/logging-operator/pkg/key"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
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

func (l *LoggingReconciler) Reconcile(ctx context.Context, lc loggedcluster.Interface) (result ctrl.Result, err error) {

	loggingEnabled := common.IsLoggingEnabled(lc)

	if loggingEnabled {
		// TODO: handle returned ctrl.Result
		_, err = l.reconcileCreate(ctx, lc)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	} else {
		_, err = l.reconcileDelete(ctx, lc)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	return ctrl.Result{}, nil
}

// reconcileCreate handles creation/update logic by calling ReconcileCreate method on all l.Reconcilers.
func (l *LoggingReconciler) reconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("LOGGING enabled")

	// Finalizer handling needs to come first.
	logger.Info(fmt.Sprintf("checking finalizer %s", key.Finalizer))
	if !controllerutil.ContainsFinalizer(lc, key.Finalizer) {
		logger.Info(fmt.Sprintf("adding finalizer %s", key.Finalizer))
		controllerutil.AddFinalizer(lc, key.Finalizer)
		err := l.Client.Update(ctx, lc.GetObject())
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	} else {
		logger.Info(fmt.Sprintf("finalizer already added"))
	}

	// Call all reconcilers ReconcileCreate methods.
	for _, reconciler := range l.Reconcilers {
		// TODO(theo): add handling for returned ctrl.Result value.
		_, err := reconciler.ReconcileCreate(ctx, lc)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	return ctrl.Result{}, nil
}

// reconcileDelete handles deletion logic by calling reconcileDelete method on all l.Reconcilers.
func (l *LoggingReconciler) reconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("LOGGING disabled")

	// Call all reconcilers ReconcileDelete methods.
	for _, reconciler := range l.Reconcilers {
		// TODO(theo): add handling for returned ctrl.Result value.
		_, err := reconciler.ReconcileDelete(ctx, lc)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	// Finalizer handling needs to come last.
	logger.Info(fmt.Sprintf("checking finalizer %s", key.Finalizer))
	if controllerutil.ContainsFinalizer(lc, key.Finalizer) {
		logger.Info(fmt.Sprintf("removing finalizer %s", key.Finalizer))
		controllerutil.RemoveFinalizer(lc, key.Finalizer)
		err := l.Client.Update(ctx, lc.GetObject())
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	} else {
		logger.Info(fmt.Sprintf("finalizer already removed"))
	}

	return ctrl.Result{}, nil
}
