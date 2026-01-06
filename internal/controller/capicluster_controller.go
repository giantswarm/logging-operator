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
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	capi "sigs.k8s.io/cluster-api/api/core/v1beta1" //nolint:staticcheck // SA1019 deprecated package
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/giantswarm/logging-operator/pkg/config"
	"github.com/giantswarm/logging-operator/pkg/key"
)

// CapiClusterReconciler reconciles a Cluster object
// This reconciler is being decommissioned - it only cleans up finalizers now.
type CapiClusterReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
	Config config.Config
}

//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters,verbs=get;list;watch
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop.
// This operator has been decommissioned - all logging configuration is now handled by observability-operator.
// The only remaining function is to clean up finalizers from existing cluster objects.
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

	logger.Info("Cleaning up logging-operator finalizer from cluster", "name", cluster.GetName())

	// Remove finalizer if present - this is the only remaining function of this operator
	return r.removeFinalizer(ctx, cluster)
}

// removeFinalizer removes the logging-operator finalizer from the cluster if present.
// This is the only remaining function - cleaning up after decommissioning the operator.
func (r *CapiClusterReconciler) removeFinalizer(ctx context.Context, cluster *capi.Cluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if !controllerutil.ContainsFinalizer(cluster, key.Finalizer) {
		logger.Info("No finalizer to remove, nothing to do")
		return ctrl.Result{}, nil
	}

	logger.Info("Removing finalizer", "finalizer", key.Finalizer)

	// We use a patch rather than an update to avoid conflicts when multiple controllers are removing their finalizer from the ClusterCR
	// We use the patch from sigs.k8s.io/cluster-api/util/patch to handle the patching without conflicts
	patchHelper, err := patch.NewHelper(cluster, r.Client)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}
	controllerutil.RemoveFinalizer(cluster, key.Finalizer)
	if err := patchHelper.Patch(ctx, cluster); err != nil {
		logger.Error(err, "Failed to remove finalizer from cluster, requeuing", "finalizer", key.Finalizer)
		return ctrl.Result{}, errors.WithStack(err)
	}
	logger.Info("Successfully removed finalizer from cluster", "finalizer", key.Finalizer)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CapiClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capi.Cluster{}).
		Complete(r)
}
