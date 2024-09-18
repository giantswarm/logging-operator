package alloysecret

import (
	"context"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"

	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconciler implements a reconciler.Interface to handle
// Alloy secret which stores sensitive configuration values.
// TODO(theo): Remove this reconciler and the whole package. It was needed to ensure the secret created previously is now deleted. This is now replaced by the secret created in the logging-secret resource.
type Reconciler struct {
	client.Client
}

// ReconcileCreate ensures Alloy secret is created with the right credentials
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	result, err := r.ReconcileDelete(ctx, lc)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	return result, nil
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
