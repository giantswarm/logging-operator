package reconciler

import (
	"context"

	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Interface provides a reconciler interface which is the controller core logic
// for reconciliation loops.
//
// An implementation can then be used by a controller to extend its capabilities.
//
// NOTE: the returned ctrl.Result is currently ignored
type Interface interface {
	ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error)

	ReconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error)
}
