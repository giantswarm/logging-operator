package resource

import (
	"context"

	capi "sigs.k8s.io/cluster-api/api/core/v1beta1" //nolint:staticcheck // SA1019 deprecated package
	ctrl "sigs.k8s.io/controller-runtime"
)

// Interface provides a resource interface which is the controller core logic
// for reconciliation loops.
//
// An implementation can then be used by a controller to extend its capabilities.
type Interface interface {
	ReconcileCreate(ctx context.Context, cluster *capi.Cluster) (ctrl.Result, error)

	ReconcileDelete(ctx context.Context, cluster *capi.Cluster) (ctrl.Result, error)
}
