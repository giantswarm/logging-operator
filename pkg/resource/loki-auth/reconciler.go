package lokiauth

import (
	"context"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	//#nosec G101
	lokiauthSecretName      = "loki-multi-tenant-proxy-auth-config"
	lokiauthSecretNamespace = "loki"
)

// Reconciler implements a reconciler.Interface to handle
// Loki auth: a secret for the Loki-multi-tenant-proxy config
type Reconciler struct {
	client.Client
}

// ReconcileCreate ensures Loki-multi-tenant map is deleted
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("lokiauth create")

	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      lokiauthSecretName,
			Namespace: lokiauthSecretNamespace,
		},
	}

	if err := r.Client.Delete(ctx, &secret); client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info("lokiauth - done")
	return ctrl.Result{}, nil
}

// ReconcileDelete - Not much to do here when a cluster is deleted
func (r *Reconciler) ReconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("lokiauth delete")

	return ctrl.Result{}, nil
}
