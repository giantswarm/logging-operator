/*
Copyright 2023.

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
	"time"

	appv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/pkg/errors"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/giantswarm/logging-operator/internal/controller/predicates"
	"github.com/giantswarm/logging-operator/pkg/common"
	"github.com/giantswarm/logging-operator/pkg/config"
	"github.com/giantswarm/logging-operator/pkg/key"
	"github.com/giantswarm/logging-operator/pkg/resource"
)

// CapiClusterReconciler reconciles a Cluster object
type CapiClusterReconciler struct {
	Client      client.Client
	Scheme      *runtime.Scheme
	Config      config.Config
	Reconcilers []resource.Interface
}

//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters,verbs=get;list;watch
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// It compares the state specified by the Cluster object against the actual
// cluster state, and then perform operations to make the cluster state reflect
// the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *CapiClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	logger := log.FromContext(ctx)

	cluster := &capi.Cluster{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, cluster)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info("Reconciling CAPI Cluster", "name", cluster.GetName())

	// Determine if logging should be enabled or disabled
	if common.IsLoggingEnabled(cluster, r.Config.EnableLoggingFlag) {
		return r.reconcileCreate(ctx, cluster)
	} else {
		return r.reconcileDelete(ctx, cluster)
	}
}

// reconcileCreate handles creation/update logic by calling ReconcileCreate method on all reconcilers.
func (r *CapiClusterReconciler) reconcileCreate(ctx context.Context, cluster *capi.Cluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("LOGGING enabled")

	if !controllerutil.ContainsFinalizer(cluster, key.Finalizer) {
		logger.Info("adding finalizer", "finalizer", key.Finalizer)

		// We use a patch rather than an update to avoid conflicts when multiple controllers are adding their finalizer to the ClusterCR
		// We use the patch from sigs.k8s.io/cluster-api/util/patch to handle the patching without conflicts
		patchHelper, err := patch.NewHelper(cluster, r.Client)
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

	loggingAgentConfig, err := common.ToggleAgents(ctx, r.Client, cluster, r.Config)
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
	for _, reconciler := range r.Reconcilers {
		result, err := reconciler.ReconcileCreate(ctx, cluster, loggingAgentConfig)
		if err != nil || !result.IsZero() {
			return result, errors.WithStack(err)
		}
	}

	return ctrl.Result{}, nil
}

// reconcileDelete handles deletion logic by calling reconcileDelete method on all reconcilers.
func (r *CapiClusterReconciler) reconcileDelete(ctx context.Context, cluster *capi.Cluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("LOGGING disabled")

	if controllerutil.ContainsFinalizer(cluster, key.Finalizer) {
		// Get the current logging agent configuration
		loggingAgentConfig, err := common.ToggleAgents(ctx, r.Client, cluster, r.Config)
		if err != nil && !apimachineryerrors.IsNotFound(err) {
			// Errors only if this is not a 404 because the apps are already deleted.
			return ctrl.Result{}, errors.WithStack(err)
		}

		// Call all reconcilers ReconcileDelete methods.
		for _, reconciler := range r.Reconcilers {
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
		patchHelper, err := patch.NewHelper(cluster, r.Client)
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

// SetupWithManager sets up the controller with the Manager.
func (r *CapiClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capi.Cluster{}).
		// This ensures we run the reconcile loop when the observability-bundle app resource version changes.
		Watches(
			&appv1alpha1.App{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, object client.Object) []reconcile.Request {
				return []reconcile.Request{
					{NamespacedName: types.NamespacedName{
						Name:      object.GetLabels()["giantswarm.io/cluster"],
						Namespace: object.GetNamespace(),
					}},
				}
			}),
			builder.WithPredicates(predicates.ObservabilityBundleAppVersionChangedPredicate{}),
		).
		Complete(r)
}
