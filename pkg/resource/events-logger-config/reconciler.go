package eventsloggerconfig

import (
	"context"
	"reflect"
	"time"

	appv1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Reconciler implements a reconciler.Interface to handle
// GrafanaAgent config: extra grafana-agent config defining what we want to retrieve.
type Reconciler struct {
	client.Client
	DefaultWorkloadClusterNamespaces []string
}

// ReconcileCreate ensures grafana-agent config is created with the right credentials
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("grafana-agent-config create")

	// Get observability bundle app metadata.
	appMeta := common.ObservabilityBundleAppMeta(lc)
	// Retrieve the app.
	var currentApp appv1.App
	if err := r.Client.Get(ctx, types.NamespacedName{Name: lc.AppConfigName("grafana-agent"), Namespace: appMeta.GetNamespace()}, &currentApp); err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("grafana-agent-config - app not found, requeuing")
			// If the app is not found we should requeue and try again later (5 minutes is the app platform default reconciliation time)
			return ctrl.Result{RequeueAfter: time.Duration(5 * time.Minute)}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Get desired config
	values, err := generateGrafanaAgentConfig(lc, r.DefaultWorkloadClusterNamespaces)
	if err != nil {
		logger.Info("grafana-agent-config - failed generating grafana-agent config!", "error", err)
		return ctrl.Result{}, errors.WithStack(err)
	}

	desiredEventsLoggerConfig := v1.ConfigMap{
		ObjectMeta: configMeta(lc),
		Data: map[string]string{
			"values": values,
		},
	}

	// Check if config already exists.
	logger.Info("grafana-agent-config - getting", "namespace", desiredEventsLoggerConfig.GetNamespace(), "name", desiredEventsLoggerConfig.GetName())
	var currentEventsLoggerConfig v1.ConfigMap
	err = r.Client.Get(ctx, types.NamespacedName{Name: desiredEventsLoggerConfig.GetName(), Namespace: desiredEventsLoggerConfig.GetNamespace()}, &currentEventsLoggerConfig)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("grafana-agent-config not found, creating")
			err = r.Client.Create(ctx, &desiredEventsLoggerConfig)
			if err != nil {
				return ctrl.Result{}, errors.WithStack(err)
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	if !needUpdate(currentEventsLoggerConfig, desiredEventsLoggerConfig) {
		logger.Info("grafana-agent-config up to date")
		return ctrl.Result{}, nil
	}

	logger.Info("grafana-agent-config - updating")
	err = r.Client.Update(ctx, &desiredEventsLoggerConfig)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info("grafana-agent-config - done")
	return ctrl.Result{}, nil
}

// ReconcileDelete ensure grafana-agent-config is deleted for the given cluster.
func (r *Reconciler) ReconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("grafana-agent-config delete")

	// Get expected configmap.
	var currentEventsLoggerConfig v1.ConfigMap
	err := r.Client.Get(ctx, types.NamespacedName{Name: getGrafanaAgentConfigName(lc), Namespace: lc.GetAppsNamespace()}, &currentEventsLoggerConfig)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("grafana-agent-config not found, stop here")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Delete configmap.
	logger.Info("grafana-agent-config deleting", "namespace", currentEventsLoggerConfig.GetNamespace(), "name", currentEventsLoggerConfig.GetName())
	err = r.Client.Delete(ctx, &currentEventsLoggerConfig)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			// Do no throw error in case it was not found, as this means
			// it was already deleted.
			logger.Info("grafana-agent-config already deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}
	logger.Info("grafana-agent-config deleted")

	return ctrl.Result{}, nil
}

// needUpdate return true if current.Data and desired.Data do not match.
func needUpdate(current, desired v1.ConfigMap) bool {
	return !reflect.DeepEqual(current.Data, desired.Data)
}
