package lokiauth

import (
	"context"
	"reflect"

	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Reconciler implements a reconciler.Interface to handle
// Promtail toggle: enable or disable Promtail in a given Cluster.
type Reconciler struct {
	client.Client
}

// ReconcileCreate ensures Loki-multi-tenant auth map is created with the right credentials
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("lokiauth create")

	// Retrieve secret containing credentials
	var lokiAuthSecret v1.Secret
	err := r.Client.Get(ctx, types.NamespacedName{Name: loggingcredentials.LoggingCredentialsSecretMeta(lc).Name, Namespace: loggingcredentials.LoggingCredentialsSecretMeta(lc).Namespace},
		&lokiAuthSecret)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Get desired secret
	desiredLokiAuthSecret, err := GenerateLokiAuthSecret(lc, &lokiAuthSecret)
	if err != nil {
		logger.Info("lokiauth - failed generating auth config!", "error", err)
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Check if auth config already exists.
	logger.Info("lokiauth - getting", "namespace", desiredLokiAuthSecret.GetNamespace(), "name", desiredLokiAuthSecret.GetName())
	var currentLokiAuthSecret v1.Secret
	err = r.Client.Get(ctx, types.NamespacedName{Name: desiredLokiAuthSecret.GetName(), Namespace: desiredLokiAuthSecret.GetNamespace()}, &currentLokiAuthSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("lokiauth not found, creating")
			err = r.Client.Create(ctx, &desiredLokiAuthSecret)
			if err != nil {
				return ctrl.Result{}, errors.WithStack(err)
			}
		} else {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	if !needUpdate(currentLokiAuthSecret, desiredLokiAuthSecret) {
		logger.Info("lokiauth up to date")
		return ctrl.Result{}, nil
	}

	logger.Info("lokiauth - updating")
	err = r.Client.Update(ctx, &desiredLokiAuthSecret)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// trigger restart of loki-multi-tenant-auth-proxy

	return ctrl.Result{}, nil
}

// ReconcileDelete - Not much to do here when a cluster is deleted
func (r *Reconciler) ReconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("lokiauth delete")

	return ctrl.Result{}, nil
}

// needUpdate return true if current.Data and desired.Data do not match.
func needUpdate(current, desired v1.Secret) bool {
	return !reflect.DeepEqual(current.Data, desired.Data)
}
