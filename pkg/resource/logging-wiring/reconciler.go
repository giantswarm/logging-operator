package loggingwiring

import (
	"context"
	"reflect"
	"time"

	appv1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/pkg/errors"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/giantswarm/logging-operator/pkg/common"
)

// Resource implements a resource.Interface to handle
// Logging wiring: set or unset the user value configmap created by
// logging-agents-toggle in the observability bundle.
type Resource struct {
	Client client.Client
	Scheme *runtime.Scheme
}

// ReconcileCreate ensure user value configmap is set in observability bundle
// for the given cluster.
func (r *Resource) ReconcileCreate(ctx context.Context, cluster *capi.Cluster, loggingAgent *common.LoggingAgent) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("logging wiring create")

	// Get observability bundle app metadata.
	appMeta := common.ObservabilityBundleAppMeta(cluster)
	// Retrieve the app.
	logger.Info("logging wiring checking app", "namespace", appMeta.GetNamespace(), "name", appMeta.GetName())
	var currentApp appv1.App
	err := r.Client.Get(ctx, types.NamespacedName{Name: appMeta.GetName(), Namespace: appMeta.GetNamespace()}, &currentApp)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("logging wiring - app not found, requeuing")
			// If the app is not found we should requeue and try again later (5 minutes is the app platform default reconciliation time)
			return ctrl.Result{RequeueAfter: time.Duration(5 * time.Minute)}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	desiredApp := wireLogging(cluster, currentApp)
	if !reflect.DeepEqual(currentApp, *desiredApp) {
		logger.Info("logging wiring updating")
		// Update the app.
		err := r.Client.Update(ctx, desiredApp)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	logger.Info("logging wiring up to date")

	return ctrl.Result{}, nil
}

// ReconcileCreate ensure user value configmap is unset in observability bundle
// for the given cluster.
func (r *Resource) ReconcileDelete(ctx context.Context, cluster *capi.Cluster, loggingAgent *common.LoggingAgent) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("logging wiring delete")

	// Get observability bundle app metadata.
	appMeta := common.ObservabilityBundleAppMeta(cluster)
	var currentApp appv1.App
	err := r.Client.Get(ctx, types.NamespacedName{Name: appMeta.GetName(), Namespace: appMeta.GetNamespace()}, &currentApp)
	if err != nil {
		// Handle case where the app is not found.
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("logging wiring - app not found, skipping deletion")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	desiredApp := unwireLogging(cluster, currentApp)
	if !reflect.DeepEqual(currentApp, *desiredApp) {
		logger.Info("logging wiring updating")
		// Update the app.
		err := r.Client.Update(ctx, desiredApp)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	logger.Info("logging wiring up to date")

	return ctrl.Result{}, nil
}

func getWiredExtraConfig(cluster *capi.Cluster) appv1.AppExtraConfig {
	observabilityBundleConfigMapMeta := common.ObservabilityBundleConfigMapMeta(cluster)
	return appv1.AppExtraConfig{
		Kind:      "configMap",
		Name:      observabilityBundleConfigMapMeta.GetName(),
		Namespace: observabilityBundleConfigMapMeta.GetNamespace(),
		Priority:  25,
	}
}

// unwireLogging unsets the extraconfig confimap in a copy of the app
func unwireLogging(cluster *capi.Cluster, currentApp appv1.App) *appv1.App {
	desiredApp := currentApp.DeepCopy()

	wiredExtraConfig := getWiredExtraConfig(cluster)
	for index, extraConfig := range currentApp.Spec.ExtraConfigs {
		if reflect.DeepEqual(extraConfig, wiredExtraConfig) {
			desiredApp.Spec.ExtraConfigs = append(currentApp.Spec.ExtraConfigs[:index], currentApp.Spec.ExtraConfigs[index+1:]...)
		}
	}

	return desiredApp
}

// wireLogging sets the extraconfig confimap in a copy of the app.
func wireLogging(cluster *capi.Cluster, currentApp appv1.App) *appv1.App {
	desiredApp := currentApp.DeepCopy()

	wiredExtraConfig := getWiredExtraConfig(cluster)
	// We check if the extra config already exists to know if we need to remove it.
	var containsWiredExtraConfig = false
	for _, extraConfig := range currentApp.Spec.ExtraConfigs {
		if reflect.DeepEqual(extraConfig, wiredExtraConfig) {
			containsWiredExtraConfig = true
		}
	}

	if !containsWiredExtraConfig {
		desiredApp.Spec.ExtraConfigs = append(desiredApp.Spec.ExtraConfigs, wiredExtraConfig)
	}

	return desiredApp
}
