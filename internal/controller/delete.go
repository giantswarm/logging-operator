package controller

import (
	"context"

	"github.com/giantswarm/logging-operator/pkg/key"
	"github.com/pkg/errors"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *ClusterReconciler) reconcileDelete(ctx context.Context, cluster capiv1beta1.Cluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("LOGGING disabled")

	// TODO(theo): logic goes here

	// Finalizer handling needs to come last
	if controllerutil.ContainsFinalizer(&cluster, key.Finalizer) {
		controllerutil.RemoveFinalizer(&cluster, key.Finalizer)
		err := r.Client.Update(ctx, &cluster)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	return ctrl.Result{}, nil
}
