package ingressauthsecret

import (
	"context"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	capi "sigs.k8s.io/cluster-api/api/core/v1beta1"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/giantswarm/logging-operator/pkg/config"
	credentials "github.com/giantswarm/logging-operator/pkg/resource/credentials"
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

// ReconcileCreate ensures loki ingress auth map is created with the right credentials on CAPI
func (r *Resource) ReconcileCreate(ctx context.Context, cluster *capi.Cluster) (ctrl.Result, error) {
	return r.createOrUpdateSecret(ctx, cluster)
}

// ReconcileDelete - Delete the loki ingress auth secret on capi
func (r *Resource) ReconcileDelete(ctx context.Context, cluster *capi.Cluster) (ctrl.Result, error) {
	return r.createOrUpdateSecret(ctx, cluster)
}

func (r *Resource) createOrUpdateSecret(ctx context.Context, cluster *capi.Cluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Retrieve currently generated credentials
	var loggingObjectKey = types.NamespacedName{
		Name:      credentials.CredentialsSecretMeta(credentials.LoggingCredentialsName, credentials.LoggingCredentialsNamespace).Name,
		Namespace: credentials.CredentialsSecretMeta(credentials.LoggingCredentialsName, credentials.LoggingCredentialsNamespace).Namespace,
	}

	loggingSecret := ingressAuthSecret(lokiIngressAuthSecretName, lokiIngressAuthSecretNamespace)

	_, err := r.generateAuthSecret(ctx, cluster, &loggingSecret, loggingObjectKey)
	if err != nil {
		logger.Error(err, "failed to generate loki ingress auth secret")
		return ctrl.Result{}, errors.WithStack(err)
	}

	if r.Config.EnableTracingFlag {
		var tracingObjectKey = types.NamespacedName{
			Name:      credentials.CredentialsSecretMeta(credentials.TracingCredentialsName, credentials.TracingCredentialsNamespace).Name,
			Namespace: credentials.CredentialsSecretMeta(credentials.TracingCredentialsName, credentials.TracingCredentialsNamespace).Namespace,
		}

		tracingSecret := ingressAuthSecret(tempoIngressAuthSecretName, tempoIngressAuthSecretNamespace)

		_, err = r.generateAuthSecret(ctx, cluster, &tracingSecret, tracingObjectKey)
		if err != nil {
			logger.Error(err, "failed to generate Tempo ingress auth secret")
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	return ctrl.Result{}, nil
}

func (r *Resource) generateAuthSecret(ctx context.Context, cluster *capi.Cluster, credentialsSecret *v1.Secret, secretKey types.NamespacedName) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var secretCredentials v1.Secret
	if err := r.Client.Get(ctx, secretKey, &secretCredentials); err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, credentialsSecret, func() error {
		// Generate ingress auth secret
		data, err := generateIngressAuthSecret(cluster, &secretCredentials)
		if err != nil {
			logger.Error(err, "failed to generate ingress auth secret")
			return errors.WithStack(err)
		}
		credentialsSecret.StringData = data

		return nil
	})
	if err != nil {
		logger.Error(err, "failed to create ingress auth secret")
		return ctrl.Result{}, errors.WithStack(err)
	}

	return ctrl.Result{}, nil
}
