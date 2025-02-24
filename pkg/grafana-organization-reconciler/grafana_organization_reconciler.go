package grafanaorganizationreconciler

import (
	"context"

	"github.com/giantswarm/logging-operator/pkg/key"
	"github.com/giantswarm/logging-operator/pkg/reconciler"
	"github.com/giantswarm/observability-operator/api/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// LoggingReconciler reconciles logging for any supported object
type GrafanaOrganizationReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	Reconcilers []reconciler.Interface
}

func (l *GrafanaOrganizationReconciler) Reconcile(ctx context.Context) (result ctrl.Result, err error) {

	return result, errors.WithStack(err)
}

// reconcileCreate handles creation/update logic by calling ReconcileCreate method on all l.Reconcilers.
func (l *GrafanaOrganizationReconciler) reconcileCreate(ctx context.Context, grafanaOrganization v1alpha1.GrafanaOrganization) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if !controllerutil.ContainsFinalizer(&grafanaOrganization, key.Finalizer) {
		logger.Info("adding finalizer", "finalizer", key.Finalizer)

		// We use a patch rather than an update to avoid conflicts when multiple controllers are adding their finalizer to the ClusterCR
		// We use the patch from sigs.k8s.io/cluster-api/util/patch to handle the patching without conflicts
		patchHelper, err := patch.NewHelper(&grafanaOrganization, l.Client)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
		controllerutil.AddFinalizer(&grafanaOrganization, key.Finalizer)
		if err := patchHelper.Patch(ctx, &grafanaOrganization); err != nil {
			logger.Error(err, "failed to add finalizer to grafana organization", "finalizer", key.Finalizer)
			return ctrl.Result{}, errors.WithStack(err)
		}
		logger.Info("successfully added finalizer to grafana organization", "finalizer", key.Finalizer)
	}

	return ctrl.Result{}, nil
}

// reconcileDelete handles deletion logic by calling reconcileDelete method on all l.Reconcilers.
func (l *GrafanaOrganizationReconciler) reconcileDelete(ctx context.Context, grafanaOrganization v1alpha1.GrafanaOrganization) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if controllerutil.ContainsFinalizer(&grafanaOrganization, key.Finalizer) {
		// We get the latest state of the object to avoid race conditions.
		// Finalizer handling needs to come last.
		logger.Info("removing finalizer", "finalizer", key.Finalizer)

		// We use a patch rather than an update to avoid conflicts when multiple controllers are removing their finalizer from the ClusterCR
		// We use the patch from sigs.k8s.io/cluster-api/util/patch to handle the patching without conflicts
		patchHelper, err := patch.NewHelper(&grafanaOrganization, l.Client)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
		controllerutil.RemoveFinalizer(&grafanaOrganization, key.Finalizer)
		if err := patchHelper.Patch(ctx, &grafanaOrganization); err != nil {
			logger.Error(err, "failed to remove finalizer from grafana organization, requeuing", "finalizer", key.Finalizer)
			return ctrl.Result{}, errors.WithStack(err)
		}
		logger.Info("successfully removed finalizer from grafana organization", "finalizer", key.Finalizer)
	}

	return ctrl.Result{}, nil
}
