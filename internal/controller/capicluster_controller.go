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

	appv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	grafanaorganization "github.com/giantswarm/observability-operator/api/v1alpha1"
	"github.com/pkg/errors"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/giantswarm/logging-operator/internal/controller/predicates"
	grafanaorganizationreconciler "github.com/giantswarm/logging-operator/pkg/grafana-organization-reconciler"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	"github.com/giantswarm/logging-operator/pkg/logged-cluster/capicluster"
	loggingreconciler "github.com/giantswarm/logging-operator/pkg/logging-reconciler"
)

// CapiClusterReconciler reconciles a Cluster object
type CapiClusterReconciler struct {
	client.Client
	Scheme                        *runtime.Scheme
	LoggingReconciler             loggingreconciler.LoggingReconciler
	GrafanaOrganizationReconciler grafanaorganizationreconciler.GrafanaOrganizationReconciler
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

	loggedCluster := &capicluster.Object{
		Object:  cluster,
		Options: loggedcluster.O,
	}

	return r.LoggingReconciler.Reconcile(ctx, loggedCluster)
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
		// This ensures we run the reconcile loop whenever there is a change in the grafana organization CRs.
		Watches(
			&grafanaorganization.GrafanaOrganization{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, object client.Object) []reconcile.Request {
				return []reconcile.Request{
					{NamespacedName: types.NamespacedName{
						Name:      object.GetLabels()["giantswarm.io/cluster"],
						Namespace: object.GetNamespace(),
					}},
				}
			}),
		).
		Complete(r)
}
