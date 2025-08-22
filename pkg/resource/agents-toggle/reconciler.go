package agentstoggle

import (
	"context"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/logging-operator/pkg/common"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Resource implements a resource.Interface to handle
// Logging agents toggle: enable or disable logging agents in a given Cluster.
type Resource struct {
	Client client.Client
	Scheme *runtime.Scheme
}

// ReconcileCreate ensure logging agents and events loggers are enabled in the given cluster.
func (r *Resource) ReconcileCreate(ctx context.Context, cluster *capi.Cluster, loggingAgent *common.LoggingAgent) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("agents toggle create")

	desiredConfigMap := v1.ConfigMap{
		ObjectMeta: common.ObservabilityBundleConfigMapMeta(cluster),
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &desiredConfigMap, func() error {
		config, err := generateObservabilityBundleConfig(loggingAgent)
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
func (r *Resource) ReconcileDelete(ctx context.Context, cluster *capi.Cluster, loggingAgent *common.LoggingAgent) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("delete agents toggle config")

	desiredConfigMap := v1.ConfigMap{
		ObjectMeta: common.ObservabilityBundleConfigMapMeta(cluster),
	}

	logger.Info("deleting agents toggle config")
	err := r.Client.Delete(ctx, &desiredConfigMap)
	if client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}
	logger.Info("agents toggle config deleted")

	return ctrl.Result{}, nil
}
