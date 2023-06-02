package controller

import (
	"context"
	"fmt"

	"github.com/giantswarm/logging-operator/pkg/key"
	"github.com/pkg/errors"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *ClusterReconciler) reconcileCreate(ctx context.Context, cluster capiv1beta1.Cluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("LOGGING enabled")

	// Finalizer handling needs to come first
	logger.Info(fmt.Sprintf("checking finalizer %s", key.Finalizer))
	if !controllerutil.ContainsFinalizer(&cluster, key.Finalizer) {
		logger.Info(fmt.Sprintf("adding finalizer %s", key.Finalizer))
		controllerutil.AddFinalizer(&cluster, key.Finalizer)
		err := r.Client.Update(ctx, &cluster)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	for _, reconciler := range r.Reconcilers {
		_, err := reconciler.ReconcileCreate(ctx, cluster)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	return ctrl.Result{}, nil
}
