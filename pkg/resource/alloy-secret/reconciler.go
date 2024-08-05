package alloysecret

import (
	"context"
	"fmt"
	"reflect"
	"time"

	appv1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconciler implements a reconciler.Interface to handle
// Alloy secret which stores sensitive configuration values.
type Reconciler struct {
	client.Client
}

// ReconcileCreate ensures Alloy secret is created with the right credentials
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("alloy-secret - create")

	if lc.GetLoggingAgent() != common.AlloyLogAgentAppName {
		logger.Info(fmt.Sprintf("alloy-secret - logging agent is not %s, skipping", common.AlloyLogAgentAppName))
		result, err := r.ReconcileDelete(ctx, lc)
		if err != nil {
			return result, errors.WithStack(err)
		}
		return result, nil
	}

	// Check existence of Alloy app
	var currentApp appv1.App
	appMeta := common.ObservabilityBundleAppMeta(lc)
	err := r.Client.Get(ctx, types.NamespacedName{Name: lc.AppConfigName(common.AlloyLogAgentAppName), Namespace: appMeta.GetNamespace()}, &currentApp)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("alloy-secret - %s app not found, requeuing", common.AlloyLogAgentAppName))
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
	lokiURL, err := common.ReadLokiIngressURL(ctx, lc, r.Client)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Get desired secret
	desiredAlloySecret, err := GenerateAlloyLoggingSecret(lc, &loggingCredentialsSecret, lokiURL)
	if err != nil {
		logger.Info("alloy-secret - failed generating alloy secret!", "error", err)
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Check if secret already exists.
	logger.Info("alloy-secret - getting secret", "namespace", desiredAlloySecret.GetNamespace(), "name", desiredAlloySecret.GetName())
	var currentAlloySecret v1.Secret
	err = r.Client.Get(ctx, types.NamespacedName{Name: desiredAlloySecret.GetName(), Namespace: desiredAlloySecret.GetNamespace()}, &currentAlloySecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("alloy-secret - secret not found, creating")
			err = r.Client.Create(ctx, &desiredAlloySecret)
			if err != nil {
				return ctrl.Result{}, errors.WithStack(err)
			}
		} else {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	if !needUpdate(currentAlloySecret, desiredAlloySecret) {
		logger.Info("alloy-secret - secret up to date")
		return ctrl.Result{}, nil
	}

	logger.Info("alloy-secret - updating secret", "namespace", desiredAlloySecret.GetNamespace(), "name", desiredAlloySecret.GetName())
	err = r.Client.Update(ctx, &desiredAlloySecret)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info("alloy-secret - secret updated")
	return ctrl.Result{}, nil
}

// ReconcileDelete ensure Alloy secret is deleted for the given cluster.
func (r *Reconciler) ReconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("alloy-secret - delete")

	// Get expected secret.
	var currentAlloySecret v1.Secret
	alloySecretMeta := SecretMeta(lc)
	alloySecretObjectKey := types.NamespacedName{Name: alloySecretMeta.GetName(), Namespace: alloySecretMeta.GetNamespace()}
	err := r.Client.Get(ctx, alloySecretObjectKey, &currentAlloySecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("alloy-secret - secret not found, stopping")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Delete secret.
	logger.Info("alloy-secret - deleting secret", "namespace", currentAlloySecret.GetNamespace(), "name", currentAlloySecret.GetName())
	err = r.Client.Delete(ctx, &currentAlloySecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			// Do no throw error in case it was not found, as this means
			// it was already deleted.
			logger.Info("alloy-secret - secret already deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info("alloy-secret - secret deleted")
	return ctrl.Result{}, nil
}

// needUpdate return true if current.Data and desired.Data do not match.
func needUpdate(current, desired v1.Secret) bool {
	return !reflect.DeepEqual(current.Data, desired.Data)
}
