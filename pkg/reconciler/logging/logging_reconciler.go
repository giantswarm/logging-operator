package logging

import (
	"context"
	"time"

	"github.com/pkg/errors"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/giantswarm/logging-operator/pkg/common"
	"github.com/giantswarm/logging-operator/pkg/config"
	"github.com/giantswarm/logging-operator/pkg/key"
	"github.com/giantswarm/logging-operator/pkg/reconciler"
)

// LoggingReconciler reconciles logging for any supported object
type LoggingReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	Reconcilers []reconciler.Interface
	Config      config.Config
}

func (l *LoggingReconciler) Reconcile(ctx context.Context, cluster *capi.Cluster) (result ctrl.Result, err error) {
	if common.IsLoggingEnabled(cluster, l.Config.EnableLoggingFlag) {
		result, err = l.reconcileCreate(ctx, cluster)
	} else {
		result, err = l.reconcileDelete(ctx, cluster)
	}

	return result, errors.WithStack(err)
}

// reconcileCreate handles creation/update logic by calling ReconcileCreate method on all l.Reconcilers.
func (l *LoggingReconciler) reconcileCreate(ctx context.Context, cluster *capi.Cluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("LOGGING enabled")

	if !controllerutil.ContainsFinalizer(cluster, key.Finalizer) {
		logger.Info("adding finalizer", "finalizer", key.Finalizer)

		// We use a patch rather than an update to avoid conflicts when multiple controllers are adding their finalizer to the ClusterCR
		// We use the patch from sigs.k8s.io/cluster-api/util/patch to handle the patching without conflicts
		patchHelper, err := patch.NewHelper(cluster, l.Client)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
		controllerutil.AddFinalizer(cluster, key.Finalizer)
		if err := patchHelper.Patch(ctx, cluster); err != nil {
			logger.Error(err, "failed to add finalizer to logger cluster", "finalizer", key.Finalizer)
			return ctrl.Result{}, errors.WithStack(err)
		}
		logger.Info("successfully added finalizer to logged cluster", "finalizer", key.Finalizer)
	}

	loggingAgentConfig, err := common.ToggleAgents(ctx, l.Client, cluster, l.Config)
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
		result, err := reconciler.ReconcileCreate(ctx, cluster, loggingAgentConfig)
		if err != nil || !result.IsZero() {
			return result, errors.WithStack(err)
		}
	}

	return ctrl.Result{}, nil
}

// reconcileDelete handles deletion logic by calling reconcileDelete method on all l.Reconcilers.
func (l *LoggingReconciler) reconcileDelete(ctx context.Context, cluster *capi.Cluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("LOGGING disabled")

	if controllerutil.ContainsFinalizer(cluster, key.Finalizer) {
		// Get the current logging agent configuration
		loggingAgentConfig, err := common.ToggleAgents(ctx, l.Client, cluster, l.Config)
		if err != nil && !apimachineryerrors.IsNotFound(err) {
			// Errors only if this is not a 404 because the apps are already deleted.
			return ctrl.Result{}, errors.WithStack(err)
		}

		// Call all reconcilers ReconcileDelete methods.
		for _, reconciler := range l.Reconcilers {
			result, err := reconciler.ReconcileDelete(ctx, cluster, loggingAgentConfig)
			if err != nil || !result.IsZero() {
				return result, errors.WithStack(err)
			}
		}

		// We get the latest state of the object to avoid race conditions.
		// Finalizer handling needs to come last.
		logger.Info("removing finalizer", "finalizer", key.Finalizer)

		// We use a patch rather than an update to avoid conflicts when multiple controllers are removing their finalizer from the ClusterCR
		// We use the patch from sigs.k8s.io/cluster-api/util/patch to handle the patching without conflicts
		patchHelper, err := patch.NewHelper(cluster, l.Client)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
		controllerutil.RemoveFinalizer(cluster, key.Finalizer)
		if err := patchHelper.Patch(ctx, cluster); err != nil {
			logger.Error(err, "failed to remove finalizer from logger cluster, requeuing", "finalizer", key.Finalizer)
			return ctrl.Result{}, errors.WithStack(err)
		}
		logger.Info("successfully removed finalizer from logged cluster", "finalizer", key.Finalizer)
	}

	return ctrl.Result{}, nil
}
