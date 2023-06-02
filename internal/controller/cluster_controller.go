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

	"github.com/giantswarm/logging-operator/pkg/reconciler"
	clusterhelpers "github.com/giantswarm/logging-operator/pkg/resource/cluster-helpers"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	Reconcilers []reconciler.Interface
}

//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters,verbs=get;list;watch
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Cluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	logger := log.FromContext(ctx)

	foundClusterCR := &capiv1beta1.Cluster{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, foundClusterCR)
	if err != nil {
		// TODO(theo): might need to ignore when objects are not found since we cannot do anything
		//             see https://book.kubebuilder.io/reference/using-finalizers.html
		//if r.Client.IsNotFound(err) {
		//	return ctrl.Result{}, nil
		//}
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info(fmt.Sprintf("Name %s", foundClusterCR.GetName()))

	// Logging should be disable in case:
	//   - logging is disabled via label
	//   - the cluster is being deleted
	disableCondition := !clusterhelpers.IsLoggingEnabled(*foundClusterCR) || !foundClusterCR.DeletionTimestamp.IsZero()

	if disableCondition {
		result, err = r.reconcileDelete(ctx, *foundClusterCR)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	} else {
		result, err = r.reconcileCreate(ctx, *foundClusterCR)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capiv1beta1.Cluster{}).
		Complete(r)
}
