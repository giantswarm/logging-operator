package eventsloggersecret

import (
	"context"
	"reflect"
	"time"

	"github.com/blang/semver"
	appv1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Reconciler implements a reconciler.Interface to handle
// grafana-agent secret: extra secret which stores logging write credentials
type Reconciler struct {
	client.Client
}

// ReconcileCreate ensures grafana-agent secret is created with the right credentials
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("grafana-agent-secret create")

	observabilityBundleVersion, err := common.GetObservabilityBundleAppVersion(lc, r.Client, ctx)
	if err != nil {
		// Handle case where the app is not found.
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("grafana-agent-secret - observability bundle app not found, requeueing")
			// If the app is not found we should requeue and try again later (5 minutes is the app platform default reconciliation time)
			return ctrl.Result{RequeueAfter: time.Duration(5 * time.Minute)}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	// The grafana agent was added only for bundle version 0.9.0 and above (cf. https://github.com/giantswarm/observability-bundle/compare/v0.8.9...v0.9.0)
	if observabilityBundleVersion.LT(semver.MustParse("0.9.0")) {
		return ctrl.Result{}, nil
	}

	// Check existence of grafana-agent app
	var currentApp appv1.App
	appMeta := common.ObservabilityBundleAppMeta(lc)
	err = r.Client.Get(ctx, types.NamespacedName{Name: lc.AppConfigName("grafana-agent"), Namespace: appMeta.GetNamespace()}, &currentApp)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("grafana-agent-secret - app not found, requeuing")
			// If the app is not found we should requeue and try again later (5 minutes is the app platform default reconciliation time)
			return ctrl.Result{RequeueAfter: time.Duration(5 * time.Minute)}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Retrieve secret containing credentials
	var loggingCredentialsSecret v1.Secret
	err = r.Client.Get(ctx, types.NamespacedName{Name: loggingcredentials.LoggingCredentialsSecretMeta(lc).Name, Namespace: loggingcredentials.LoggingCredentialsSecretMeta(lc).Namespace},
		&loggingCredentialsSecret)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Retrieve Loki ingress name
	lokiURL, err := common.ReadProxyIngressURL(ctx, lc, r.Client)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Get desired secret
	desiredGrafanaAgentSecret, err := GenerateGrafanaAgentSecret(lc, &loggingCredentialsSecret, lokiURL)
	if err != nil {
		logger.Info("grafana-agent-secret - failed generating auth config!", "error", err)
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Check if secret already exists.
	logger.Info("grafana-agent-secret - getting", "namespace", desiredGrafanaAgentSecret.GetNamespace(), "name", desiredGrafanaAgentSecret.GetName())
	var currentGrafanaAgentSecret v1.Secret
	err = r.Client.Get(ctx, types.NamespacedName{Name: desiredGrafanaAgentSecret.GetName(), Namespace: desiredGrafanaAgentSecret.GetNamespace()}, &currentGrafanaAgentSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("grafana-agent-secret not found, creating")
			err = r.Client.Create(ctx, &desiredGrafanaAgentSecret)
			if err != nil {
				return ctrl.Result{}, errors.WithStack(err)
			}
		} else {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	if !needUpdate(currentGrafanaAgentSecret, desiredGrafanaAgentSecret) {
		logger.Info("grafana-agent-secret up to date")
		return ctrl.Result{}, nil
	}

	logger.Info("grafana-agent-secret - updating")
	err = r.Client.Update(ctx, &desiredGrafanaAgentSecret)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info("grafana-agent-secret - done")
	return ctrl.Result{}, nil
}

// ReconcileDelete ensure grafana-agent-secret is deleted for the given cluster.
func (r *Reconciler) ReconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("grafana-agent-secret delete")

	// Get expected secret.
	var currentGrafanaAgentSecret v1.Secret
	err := r.Client.Get(ctx, types.NamespacedName{Name: getGrafanaAgentSecretName(lc), Namespace: lc.GetAppsNamespace()}, &currentGrafanaAgentSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("grafana-agent-secret not found, stop here")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Delete secret.
	logger.Info("grafana-agent-secret deleting", "namespace", currentGrafanaAgentSecret.GetNamespace(), "name", currentGrafanaAgentSecret.GetName())
	err = r.Client.Delete(ctx, &currentGrafanaAgentSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			// Do no throw error in case it was not found, as this means
			// it was already deleted.
			logger.Info("grafana-agent-secret already deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}
	logger.Info("grafana-agent-secret deleted")

	return ctrl.Result{}, nil
}

// needUpdate return true if current.Data and desired.Data do not match.
func needUpdate(current, desired v1.Secret) bool {
	return !reflect.DeepEqual(current.Data, desired.Data)
}
