package reconciler

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Interface provides a reconciler interface which is the controller core logic
// for reconciliation loops.
//
// An implementation can then be used by a controller to extend its capabilities.
//
// NOTE: the returned ctrl.Result is currently ignored
type Interface interface {
	ReconcileCreate(ctx context.Context, object client.Object) (ctrl.Result, error)

	ReconcileDelete(ctx context.Context, object client.Object) (ctrl.Result, error)
}
