package reconciler

import (
	"context"

	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Interface interface {
	ReconcileCreate(ctx context.Context, cluster capiv1beta1.Cluster) (ctrl.Result, error)

	ReconcileDelete(ctx context.Context, cluster capiv1beta1.Cluster) (ctrl.Result, error)
}
