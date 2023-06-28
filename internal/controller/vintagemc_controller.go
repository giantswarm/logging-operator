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

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	loggingreconciler "github.com/giantswarm/logging-operator/pkg/logging-reconciler"
	"github.com/giantswarm/logging-operator/pkg/vintagemc"
)

// VintageMCReconciler reconciles a Service object
type VintageMCReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	LoggingReconciler loggingreconciler.LoggingReconciler
}

//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=services/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Service object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
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
		// TODO(theo): might need to ignore when objects are not found since we cannot do anything
		//             see https://book.kubebuilder.io/reference/using-finalizers.html
		//if r.Client.IsNotFound(err) {
		//  return ctrl.Result{}, nil
		//}
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info("Reconciling Management cluster")

	// TODO(theo): Pass the IsLoggingEnabled function as a parameter into the LoggingReconciler
	// So we can have different detection logic to enable logging for Vintage MC and CAPI cluster.
	// On Vintage MC we determine if logging is enabled based on a global installation
	// level setting which need to be passed as a flag to this operator.
	// IsLoggingEnabled for this controller would only check the given flag, while the CAPI controller
	// would check both the flag and the label.

	loggedCluster := vintagemc.Object{
		Object: service,
	}
	result, err = r.LoggingReconciler.Reconcile(ctx, loggedCluster)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VintageMCReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		Complete(r)
}
