package lokiingressauthsecret

import (
	"context"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Reconciler implements a reconciler.Interface to handle
// loki ingress auth secret: a secret for the loki ingress that adds support for basic auth for the write path
type Reconciler struct {
	client.Client
}

// ReconcileCreate ensures loki ingress auth map is created with the right credentials on CAPI
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	// If we are not on CAPI, we don't need to create the secret as we are using the multi-tenant-proxy
	if !lc.IsCAPI() {
		return ctrl.Result{}, nil
	}

	return r.createOrUpdateSecret(ctx, lc)
}

// ReconcileDelete - Delete the loki ingress auth secret on capi
func (r *Reconciler) ReconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	// If we are not on CAPI, we don't need to create the secret as we are using the multi-tenant-proxy
	if !lc.IsCAPI() {
		return ctrl.Result{}, nil
	}

	return r.createOrUpdateSecret(ctx, lc)
}

func (r *Reconciler) createOrUpdateSecret(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Retrieve currently generated write path credentials
	var objectKey = types.NamespacedName{
		Name:      loggingcredentials.LoggingCredentialsSecretMeta().Name,
		Namespace: loggingcredentials.LoggingCredentialsSecretMeta().Namespace,
	}

	var writePathCredentials v1.Secret
	if err := r.Client.Get(ctx, objectKey, &writePathCredentials); err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	secret := lokiIngressAuthSecret()
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &secret, func() error {
		// Generate loki ingress auth secret
		data, err := generateLokiIngressAuthSecret(lc, &writePathCredentials)
		if err != nil {
			logger.Error(err, "failed to generate loki ingress auth secret")
			return errors.WithStack(err)
		}
		secret.Data = data

		return nil
	})
	if err != nil {
		logger.Error(err, "failed to create loki ingress auth secret")
		return ctrl.Result{}, errors.WithStack(err)
	}

	return ctrl.Result{}, nil
}
