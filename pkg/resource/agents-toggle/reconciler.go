package agentstoggle

import (
	"context"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	capi "sigs.k8s.io/cluster-api/api/core/v1beta1" //nolint:staticcheck // SA1019 deprecated package

	"github.com/giantswarm/logging-operator/pkg/common"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Resource implements a resource.Interface to handle
// Logging agents toggle: enable or disable logging agents in a given Cluster.
type Resource struct {
	Client client.Client
	Scheme *runtime.Scheme
}

// ReconcileCreate ensure logging agents and events loggers are enabled in the given cluster.
func (r *Resource) ReconcileCreate(ctx context.Context, cluster *capi.Cluster) (ctrl.Result, error) {
	return r.ReconcileDelete(ctx, cluster)
}

// ReconcileDelete ensure logging agents and events loggers are disabled for the given cluster.
func (r *Resource) ReconcileDelete(ctx context.Context, cluster *capi.Cluster) (ctrl.Result, error) {
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
