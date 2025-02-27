/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	grafanaorganization "github.com/giantswarm/observability-operator/api/v1alpha1"
	"github.com/pkg/errors"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/giantswarm/logging-operator/pkg/key"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	"github.com/giantswarm/logging-operator/pkg/logged-cluster/capicluster"
	loggingconfig "github.com/giantswarm/logging-operator/pkg/resource/logging-config"
)

// GrafanaOrganizationReconciler reconciles grafanaOrganization CRs
type GrafanaOrganizationReconciler struct {
	client.Client
	Scheme                  *runtime.Scheme
	LoggingConfigReconciler loggingconfig.Reconciler
}

//+kubebuilder:rbac:groups=observability.giantswarm.io,resources=grafanaorganizations,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=observability.giantswarm.io,resources=grafanaorganizations/finalizers,verbs=update

func (g *GrafanaOrganizationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	logger := log.FromContext(ctx)

	logger.Info("Started reconciling Grafana Organization")
	defer logger.Info("Finished reconciling Grafana Organization")

	grafanaOrganization := &grafanaorganization.GrafanaOrganization{}
	err = g.Client.Get(ctx, req.NamespacedName, grafanaOrganization)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(client.IgnoreNotFound(err))
	}

	clusters := &capi.ClusterList{}
	err = g.Client.List(ctx, clusters)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	for _, cluster := range clusters.Items {
		loggedCluster := &capicluster.Object{
			Object:  &cluster,
			Options: loggedcluster.O,
		}

		logger.Info("logged cluster", "name", loggedCluster)

		// Handle deleted grafana organizations
		if !grafanaOrganization.DeletionTimestamp.IsZero() {
			return ctrl.Result{}, g.reconcileDelete(ctx, *grafanaOrganization, loggedCluster)
		} else {
			// Handle non-deleted grafana organizations
			return g.reconcileCreate(ctx, *grafanaOrganization, loggedCluster)
		}
	}

	return ctrl.Result{}, nil
}

func (g *GrafanaOrganizationReconciler) reconcileCreate(ctx context.Context, grafanaOrganization grafanaorganization.GrafanaOrganization, lc loggedcluster.Interface) (ctrl.Result, error) {
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

func (g *GrafanaOrganizationReconciler) reconcileDelete(ctx context.Context, grafanaOrganization grafanaorganization.GrafanaOrganization, lc loggedcluster.Interface) error {
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

// SetupWithManager sets up the controller with the Manager.
func (g *GrafanaOrganizationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&grafanaorganization.GrafanaOrganization{}).
		Complete(g)
}
