package reconciler

import (
	"context"

	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Interface provides a reconciler interface which is the controller core logic
// for reconciliation loops.
//
// An implementation can then be used by a controller to extend its capabilities.
//
// NOTE: the returned ctrl.Result is currently ignored
type Interface interface {
	ReconcileCreate(ctx context.Context, cluster capiv1beta1.Cluster) (ctrl.Result, error)

	ReconcileDelete(ctx context.Context, cluster capiv1beta1.Cluster) (ctrl.Result, error)
}
