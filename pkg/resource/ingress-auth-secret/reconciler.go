package ingressauthsecret

import (
	"context"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	capi "sigs.k8s.io/cluster-api/api/core/v1beta1" //nolint:staticcheck // SA1019 deprecated package

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/giantswarm/logging-operator/pkg/config"
)

const (
	lokiIngressAuthSecretName      = "loki-ingress-auth" //#nosec G101
	lokiIngressAuthSecretNamespace = "loki"
	// The Tempo Ingress resource is defined here: https://github.com/giantswarm/shared-configs/blob/main/default/apps/tempo/configmap-values.yaml.template#L9
	tempoIngressAuthSecretName      = "tempo-ingress-auth" //#nosec G101
	tempoIngressAuthSecretNamespace = "tempo"
)

// Resource implements a resource.Interface to handle
// loki ingress auth secret: a secret for the loki ingress that adds support for basic auth for the write path
type Resource struct {
	Client client.Client
	Config config.Config
}

// ReconcileCreate always calls cleanup to remove old ingress auth secrets.
// Auth secrets are now managed by observability-operator's auth manager.
func (r *Resource) ReconcileCreate(ctx context.Context, _ *capi.Cluster) (ctrl.Result, error) {
	return r.deleteAuthSecrets(ctx)
}

// ReconcileDelete - Delete the loki ingress auth secret on capi
func (r *Resource) ReconcileDelete(ctx context.Context, _ *capi.Cluster) (ctrl.Result, error) {
	return r.deleteAuthSecrets(ctx)
}

func (r *Resource) deleteAuthSecrets(ctx context.Context) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	loggingSecret := ingressAuthSecret(lokiIngressAuthSecretName, lokiIngressAuthSecretNamespace)

	_, err := r.deleteSecret(ctx, &loggingSecret)
	if err != nil {
		logger.Error(err, "failed to generate loki ingress auth secret")
		return ctrl.Result{}, errors.WithStack(err)
	}

	if r.Config.EnableTracingFlag {
		tracingSecret := ingressAuthSecret(tempoIngressAuthSecretName, tempoIngressAuthSecretNamespace)
		_, err = r.deleteSecret(ctx, &tracingSecret)
		if err != nil {
			logger.Error(err, "failed to generate Tempo ingress auth secret")
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	return ctrl.Result{}, nil
}

func (r *Resource) deleteSecret(ctx context.Context, credentialsSecret *v1.Secret) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	err := r.Client.Delete(ctx, credentialsSecret)
	if client.IgnoreNotFound(err) != nil {
		logger.Error(err, "failed to delete ingress auth secret")
		return ctrl.Result{}, errors.WithStack(err)
	}

	return ctrl.Result{}, nil
}
