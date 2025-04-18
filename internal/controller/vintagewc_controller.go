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
	"fmt"

	appv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/pkg/errors"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/giantswarm/logging-operator/internal/controller/predicates"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	"github.com/giantswarm/logging-operator/pkg/logged-cluster/vintagewc"
	"github.com/giantswarm/logging-operator/pkg/reconciler/logging"
)

// VintageWCReconciler reconciles a Cluster object
type VintageWCReconciler struct {
	Client     client.Client
	Scheme     *runtime.Scheme
	Reconciler logging.LoggingReconciler
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
func (r *VintageWCReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	logger := log.FromContext(ctx)

	cluster := &capiv1beta1.Cluster{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, cluster)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info("Reconciling Vintage WC cluster", "name", cluster.GetName())

	loggedCluster := &vintagewc.Object{
		Object:  cluster,
		Options: loggedcluster.O,
		LoggingAgent: &loggedcluster.LoggingAgent{
			LoggingAgent:     loggedcluster.O.DefaultLoggingAgent,
			KubeEventsLogger: loggedcluster.O.DefaultKubeEventsLogger,
		},
	}
	return r.Reconciler.Reconcile(ctx, loggedCluster)
}

// SetupWithManager sets up the controller with the Manager.
func (r *VintageWCReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capiv1beta1.Cluster{}).
		// This ensures we run the reconcile loop when the observability-bundle app resource version changes.
		Watches(
			&appv1alpha1.App{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, object client.Object) []reconcile.Request {
				return []reconcile.Request{
					{NamespacedName: types.NamespacedName{
						Name:      object.GetLabels()["giantswarm.io/cluster"],
						Namespace: fmt.Sprintf("org-%s", object.GetLabels()["giantswarm.io/organization"]),
					}},
				}
			}),
			builder.WithPredicates(predicates.ObservabilityBundleAppVersionChangedPredicate{}),
		).
		Complete(r)
}
