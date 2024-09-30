package podlogs

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Reconciler implements a reconciler.Interface.
// This reconciler is responsible for ensure PodLogs resources are created/deleted when appropriate.
type Reconciler struct {
	client.Client
}

// ReconcileCreate ensures PodLogs is created when using Alloy as logging agent.
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("podlogs - create")

	if lc.GetLoggingAgent() != common.LoggingAgentAlloy {
		result, err := r.ReconcileDelete(ctx, lc)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}

		return result, nil
	}

	podLogGetter := PodLog()
	podLog := podLogGetter.GetWithMetaOnly()
	result, err := controllerutil.CreateOrUpdate(ctx, r.Client, podLog, func() error {
		podLog.Spec = podLogGetter.GetSpec()
		return nil
	})
	if err != nil {
		logger.WithValues("podlogs", podLog.GetName()).Error(err, "podlogs - create failed")
		return ctrl.Result{}, errors.WithStack(err)
	}
	logger.WithValues("podlogs", podLog.GetName()).Info(fmt.Sprintf("podlogs - create result: %v", result))

	logger.Info("podlogs - created")
	return ctrl.Result{}, nil
}

// ReconcileDelete ensures PodLogs is deleted.
func (r *Reconciler) ReconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("podlogs - delete")

	podLogGetter := PodLog()
	podLog := podLogGetter.GetWithMetaOnly()
	err := r.Client.Delete(ctx, podLog)
	if client.IgnoreNotFound(err) != nil {
		logger.WithValues("podlogs", podLog.GetName()).Error(err, "podlogs - delete failed")
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info("podlogs - deleted")
	return ctrl.Result{}, nil
}