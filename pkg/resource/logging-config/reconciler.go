package loggingconfig

import (
	"context"
	"reflect"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ollyop "github.com/giantswarm/observability-operator/pkg/common/tenancy"

	"github.com/giantswarm/logging-operator/pkg/common"
	"github.com/giantswarm/logging-operator/pkg/config"
)

// Resource implements a resource.Interface to handle
// Logging config: extra logging config defining what we want to retrieve.
type Resource struct {
	Client                           client.Client
	Config                           config.Config
	DefaultWorkloadClusterNamespaces []string
}

// ReconcileCreate ensures logging-config is created with the right credentials
func (r *Resource) ReconcileCreate(ctx context.Context, cluster *capi.Cluster, loggingAgent *common.LoggingAgent) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("logging-config create")

	observabilityBundleVersion, err := common.GetObservabilityBundleAppVersion(ctx, r.Client, cluster)
	if err != nil {
		// Handle case where the app is not found.
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("logging-config - observability bundle app not found, requeueing")
			// If the app is not found we should requeue and try again later (5 minutes is the app platform default reconciliation time)
			return ctrl.Result{RequeueAfter: time.Duration(5 * time.Minute)}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Get list of tenants
	tenants, err := ollyop.ListTenants(ctx, r.Client)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Get desired config
	desiredLoggingConfig, err := GenerateLoggingConfig(cluster, loggingAgent, observabilityBundleVersion, r.DefaultWorkloadClusterNamespaces, tenants, r.Config.InstallationName, r.Config.InsecureCA)
	if err != nil {
		logger.Info("logging-config - failed generating logging config!", "error", err)
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Check if config already exists.
	logger.Info("logging-config - getting", "namespace", desiredLoggingConfig.GetNamespace(), "name", desiredLoggingConfig.GetName())
	var currentLoggingConfig v1.ConfigMap
	err = r.Client.Get(ctx, types.NamespacedName{Name: desiredLoggingConfig.GetName(), Namespace: desiredLoggingConfig.GetNamespace()}, &currentLoggingConfig)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("logging-config not found, creating")
			err = r.Client.Create(ctx, &desiredLoggingConfig)
			if err != nil {
				return ctrl.Result{}, errors.WithStack(err)
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	if !needUpdate(currentLoggingConfig, desiredLoggingConfig) {
		logger.Info("logging-config up to date")
		return ctrl.Result{}, nil
	}

	logger.Info("logging-config - updating")
	err = r.Client.Update(ctx, &desiredLoggingConfig)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info("logging-config - done")
	return ctrl.Result{}, nil
}

// ReconcileDelete ensure logging-config is deleted for the given cluster.
func (r *Resource) ReconcileDelete(ctx context.Context, cluster *capi.Cluster, loggingAgent *common.LoggingAgent) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("logging-config delete")

	// Get expected configmap.
	var currentLoggingConfig v1.ConfigMap
	err := r.Client.Get(ctx, types.NamespacedName{Name: getLoggingConfigName(cluster), Namespace: cluster.GetNamespace()}, &currentLoggingConfig)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("logging-config not found, nothing to delete")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Delete configmap.
	logger.Info("logging-config deleting", "namespace", currentLoggingConfig.GetNamespace(), "name", currentLoggingConfig.GetName())
	err = r.Client.Delete(ctx, &currentLoggingConfig)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			// Do no throw error in case it was not found, as this means
			// it was already deleted.
			logger.Info("logging-config already deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}
	logger.Info("logging-config deleted")

	return ctrl.Result{}, nil
}

// needUpdate return true if current.Data and desired.Data do not match.
func needUpdate(current, desired v1.ConfigMap) bool {
	return !reflect.DeepEqual(current.Data, desired.Data)
}
