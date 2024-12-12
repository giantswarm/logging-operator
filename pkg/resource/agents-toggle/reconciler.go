package agentstoggle

import (
	"context"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Reconciler implements a reconciler.Interface to handle
// Logging agents toggle: enable or disable logging agents in a given Cluster.
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// ReconcileCreate ensure logging agents and events loggers are enabled in the given cluster.
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("agents toggle create")

	observabilityBundleVersion, err := common.GetObservabilityBundleAppVersion(lc, r.Client, ctx)
	if err != nil {
		// Handle case where the app is not found.
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("agents-toggle - observability bundle app not found, requeueing")
			// If the app is not found we should requeue and try again later (5 minutes is the app platform default reconciliation time)
			return ctrl.Result{RequeueAfter: time.Duration(5 * time.Minute)}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	desiredConfigMap := v1.ConfigMap{
		ObjectMeta: common.ObservabilityBundleConfigMapMeta(lc),
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, &desiredConfigMap, func() error {
		config, err := generateObservabilityBundleConfig(ctx, lc, observabilityBundleVersion)
		if err != nil {
			return errors.WithStack(err)
		}

		desiredConfigMap.Data = map[string]string{"values": config}
		return nil
	})
	if err != nil {
		logger.Error(err, "failed to toggle logging agents")
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info("agents toggle up to date")

	return ctrl.Result{}, nil
}

// ReconcileDelete ensure logging agents and events loggers are disabled for the given cluster.
func (r *Reconciler) ReconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("delete agents toggle config")

	desiredConfigMap := v1.ConfigMap{
		ObjectMeta: common.ObservabilityBundleConfigMapMeta(lc),
	}

	logger.Info("deleting agents toggle config")
	err := r.Client.Delete(ctx, &desiredConfigMap)
	if client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}
	logger.Info("agents toggle config deleted")

	return ctrl.Result{}, nil
}
