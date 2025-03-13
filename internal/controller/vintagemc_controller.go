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
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/giantswarm/logging-operator/internal/controller/predicates"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	"github.com/giantswarm/logging-operator/pkg/logged-cluster/vintagemc"
	"github.com/giantswarm/logging-operator/pkg/reconciler/logging"
)

// VintageMCReconciler reconciles a Service object
type VintageMCReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Reconciler logging.LoggingReconciler
}

//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=services/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *VintageMCReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	logger := log.FromContext(ctx)

	// We only want to watch 1 service, kubernetes in default NS.
	if req.Namespace != "default" || req.Name != "kubernetes" {
		return ctrl.Result{}, nil
	}

	service := &corev1.Service{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, service)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info("Reconciling Management cluster")

	loggedCluster := &vintagemc.Object{
		Object:  service,
		Options: loggedcluster.O,
	}
	return r.Reconciler.Reconcile(ctx, loggedCluster)
}

// SetupWithManager sets up the controller with the Manager.
func (r *VintageMCReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		// This ensures we run the reconcile loop when the observability-bundle app resource version changes.
		Watches(
			&appv1alpha1.App{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, object client.Object) []reconcile.Request {
				return []reconcile.Request{
					{NamespacedName: types.NamespacedName{
						Name:      "kubernetes",
						Namespace: "default",
					}},
				}
			}),
			builder.WithPredicates(predicates.ObservabilityBundleAppVersionChangedPredicate{}),
		).
		Complete(r)
}
