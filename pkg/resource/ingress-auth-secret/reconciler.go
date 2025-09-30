package ingressauthsecret

import (
	"context"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"
)

const (
	//#nosec G101
	lokiIngressAuthSecretName       = "loki-ingress-auth"
	lokiIngressAuthSecretNamespace  = "loki"
	tempoIngressAuthSecretName      = "tempo-ingress-auth"
	tempoIngressAuthSecretNamespace = "tempo"
)

// Resource implements a resource.Interface to handle
// loki ingress auth secret: a secret for the loki ingress that adds support for basic auth for the write path
type Resource struct {
	Client client.Client
}

// ReconcileCreate ensures loki ingress auth map is created with the right credentials on CAPI
func (r *Resource) ReconcileCreate(ctx context.Context, cluster *capi.Cluster, loggingAgent *common.LoggingAgent) (ctrl.Result, error) {
	return r.createOrUpdateSecret(ctx, cluster)
}

// ReconcileDelete - Delete the loki ingress auth secret on capi
func (r *Resource) ReconcileDelete(ctx context.Context, cluster *capi.Cluster, loggingAgent *common.LoggingAgent) (ctrl.Result, error) {
	return r.createOrUpdateSecret(ctx, cluster)
}

func (r *Resource) createOrUpdateSecret(ctx context.Context, cluster *capi.Cluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Retrieve currently generated credentials
	var loggingObjectKey = types.NamespacedName{
		Name:      loggingcredentials.LoggingCredentialsSecretMeta().Name,
		Namespace: loggingcredentials.LoggingCredentialsSecretMeta().Namespace,
	}

	var tracingObjectKey = types.NamespacedName{
		Name:      loggingcredentials.TracingCredentialsSecretMeta().Name,
		Namespace: loggingcredentials.TracingCredentialsSecretMeta().Namespace,
	}

	var loggingCredentials v1.Secret
	if err := r.Client.Get(ctx, loggingObjectKey, &loggingCredentials); err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	var tracingCredentials v1.Secret
	if err := r.Client.Get(ctx, tracingObjectKey, &tracingCredentials); err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	loggingSecret := ingressAuthSecret(lokiIngressAuthSecretName, lokiIngressAuthSecretNamespace)
	tracingSecret := ingressAuthSecret(tempoIngressAuthSecretName, tempoIngressAuthSecretNamespace)

	for i, secret := range []v1.Secret{loggingSecret, tracingSecret} {
		_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &secret, func() error {
			var credentials v1.Secret

			if i == 0 {
				credentials = loggingCredentials
			} else {
				credentials = tracingCredentials
			}

			// Generate loki ingress auth secret
			data, err := generateIngressAuthSecret(cluster, &credentials)
			if err != nil {
				logger.Error(err, "failed to generate loki ingress auth secret")
				return errors.WithStack(err)
			}
			secret.StringData = data

			return nil
		})
		if err != nil {
			logger.Error(err, "failed to create loki ingress auth secret")
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	return ctrl.Result{}, nil
}
