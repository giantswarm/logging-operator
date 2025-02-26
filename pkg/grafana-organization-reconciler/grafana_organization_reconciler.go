package grafanaorganizationreconciler

import (
	"context"

	"github.com/giantswarm/observability-operator/api/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/giantswarm/logging-operator/pkg/key"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	loggingconfig "github.com/giantswarm/logging-operator/pkg/resource/logging-config"
)

// GrafanaOrganizationReconciler reconciles grafanaOrganization CRs
type GrafanaOrganizationReconciler struct {
	client.Client
	Scheme                  *runtime.Scheme
	LoggingConfigReconciler loggingconfig.Reconciler
}

func (g *GrafanaOrganizationReconciler) Reconcile(ctx context.Context, req ctrl.Request, lc loggedcluster.Interface) (result ctrl.Result, err error) {
	logger := log.FromContext(ctx)

	logger.Info("Started reconciling Grafana Organization")
	defer logger.Info("Finished reconciling Grafana Organization")

	grafanaOrganization := &v1alpha1.GrafanaOrganization{}
	err = g.Client.Get(ctx, req.NamespacedName, grafanaOrganization)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(client.IgnoreNotFound(err))
	}

	// Handle deleted grafana organizations
	if !grafanaOrganization.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, g.reconcileDelete(ctx, *grafanaOrganization, lc)
	}

	// Handle non-deleted grafana organizations
	return g.reconcileCreate(ctx, *grafanaOrganization, lc)
}

func (g *GrafanaOrganizationReconciler) reconcileCreate(ctx context.Context, grafanaOrganization v1alpha1.GrafanaOrganization, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if !controllerutil.ContainsFinalizer(&grafanaOrganization, key.Finalizer) {
		logger.Info("adding finalizer to Grafana Organization", "finalizer", key.Finalizer)

		// We use a patch rather than an update to avoid conflicts when multiple controllers are adding their finalizer to the GrafanaOrganization
		// We use the patch from sigs.k8s.io/cluster-api/util/patch to handle the patching without conflicts
		patchHelper, err := patch.NewHelper(&grafanaOrganization, g.Client)
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

	result, err := g.LoggingConfigReconciler.ReconcileCreate(ctx, lc)
	if err != nil {
		return result, errors.WithStack(err)
	}

	return result, nil
}

func (g *GrafanaOrganizationReconciler) reconcileDelete(ctx context.Context, grafanaOrganization v1alpha1.GrafanaOrganization, lc loggedcluster.Interface) error {
	logger := log.FromContext(ctx)

	if controllerutil.ContainsFinalizer(&grafanaOrganization, key.Finalizer) {
		// We get the latest state of the object to avoid race conditions.
		// Finalizer handling needs to come last.
		logger.Info("removing finalizer from Grafana Organization", "finalizer", key.Finalizer)

		// We use a patch rather than an update to avoid conflicts when multiple controllers are removing their finalizer from the GrafanaOrganization
		// We use the patch from sigs.k8s.io/cluster-api/util/patch to handle the patching without conflicts
		patchHelper, err := patch.NewHelper(&grafanaOrganization, g.Client)
		if err != nil {
			return errors.WithStack(err)
		}
		controllerutil.RemoveFinalizer(&grafanaOrganization, key.Finalizer)
		if err := patchHelper.Patch(ctx, &grafanaOrganization); err != nil {
			logger.Error(err, "failed to remove finalizer from grafana organization, requeuing", "finalizer", key.Finalizer)
			return errors.WithStack(err)
		}
		logger.Info("successfully removed finalizer from grafana organization", "finalizer", key.Finalizer)
	}

	_, err := g.LoggingConfigReconciler.ReconcileDelete(ctx, lc)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
