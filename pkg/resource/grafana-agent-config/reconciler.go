package grafanaagentconfig

import (
	"context"
	"reflect"
	"time"

	"github.com/blang/semver"
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
}

// ReconcileCreate ensures grafana-agent config is created with the right credentials
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("grafana-agent-config create")

	observabilityBundleVersion, err := common.GetObservabilityBundleAppVersion(lc, r.Client, ctx)
	if err != nil {
		// Handle case where the app is not found.
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("grafana-agent-config - observability bundle app not found, requeueing")
			// If the app is not found we should requeue and try again later (5 minutes is the app platform default reconciliation time)
			return ctrl.Result{RequeueAfter: time.Duration(5 * time.Minute)}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	// The grafana agent was added only for bundle version 0.9.0 and above (cf. https://github.com/giantswarm/observability-bundle/compare/v0.8.9...v0.9.0)
	if observabilityBundleVersion.LT(semver.MustParse("0.9.0")) {
		return ctrl.Result{}, nil
	}

	// Retrieve the app.
	currentApp, err := common.ReadGrafanaAgentApp(ctx, lc, r.Client)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("grafana-agent-config - app not found, requeuing")
			// If the app is not found we should requeue and try again later (5 minutes is the app platform default reconciliation time)
			return ctrl.Result{RequeueAfter: time.Duration(5 * time.Minute)}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Get desired config
	desiredGrafanaAgentConfig, err := GenerateGrafanaAgentConfig(lc, currentApp.Spec.Namespace)
	if err != nil {
		logger.Info("grafana-agent-config - failed generating grafana-agent config!", "error", err)
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Check if config already exists.
	logger.Info("grafana-agent-config - getting", "namespace", desiredGrafanaAgentConfig.GetNamespace(), "name", desiredGrafanaAgentConfig.GetName())
	var currentGrafanaAgentConfig v1.ConfigMap
	err = r.Client.Get(ctx, types.NamespacedName{Name: desiredGrafanaAgentConfig.GetName(), Namespace: desiredGrafanaAgentConfig.GetNamespace()}, &currentGrafanaAgentConfig)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("grafana-agent-config not found, creating")
			err = r.Client.Create(ctx, &desiredGrafanaAgentConfig)
			if err != nil {
				return ctrl.Result{}, errors.WithStack(err)
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	if !needUpdate(currentGrafanaAgentConfig, desiredGrafanaAgentConfig) {
		logger.Info("grafana-agent-config up to date")
		return ctrl.Result{}, nil
	}

	logger.Info("grafana-agent-config - updating")
	err = r.Client.Update(ctx, &desiredGrafanaAgentConfig)
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
	var currentGrafanaAgentConfig v1.ConfigMap
	err := r.Client.Get(ctx, types.NamespacedName{Name: getGrafanaAgentConfigName(lc), Namespace: lc.GetAppsNamespace()}, &currentGrafanaAgentConfig)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("grafana-agent-config not found, stop here")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Delete configmap.
	logger.Info("grafana-agent-config deleting", "namespace", currentGrafanaAgentConfig.GetNamespace(), "name", currentGrafanaAgentConfig.GetName())
	err = r.Client.Delete(ctx, &currentGrafanaAgentConfig)
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
