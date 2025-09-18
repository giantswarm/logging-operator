package tracingsecret

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

// Resource implements a resource.Interface to handle
// tempo ingress auth secret: a secret for the tempo ingress that adds support for basic auth for the write path
type Resource struct {
	Client client.Client
}

// ReconcileCreate ensures tempo ingress auth map is created with the right credentials on CAPI
func (r *Resource) ReconcileCreate(ctx context.Context, cluster *capi.Cluster, loggingAgent *common.LoggingAgent) (ctrl.Result, error) {
	return r.createOrUpdateSecret(ctx, cluster)
}

// ReconcileDelete - Delete the tempo ingress auth secret on capi
func (r *Resource) ReconcileDelete(ctx context.Context, cluster *capi.Cluster, loggingAgent *common.LoggingAgent) (ctrl.Result, error) {
	return r.createOrUpdateSecret(ctx, cluster)
}

func (r *Resource) createOrUpdateSecret(ctx context.Context, cluster *capi.Cluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Retrieve currently generated write path credentials
	var objectKey = types.NamespacedName{
		Name:      loggingcredentials.TracingCredentialsSecretMeta().Name,
		Namespace: loggingcredentials.TracingCredentialsSecretMeta().Namespace,
	}

	var writePathCredentials v1.Secret
	if err := r.Client.Get(ctx, objectKey, &writePathCredentials); err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	secret := tempoIngressAuthSecret()
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &secret, func() error {
		// Generate tempo ingress auth secret
		data, err := generatetempoIngressAuthSecret(cluster, &writePathCredentials)
		if err != nil {
			logger.Error(err, "failed to generate tempo ingress auth secret")
			return errors.WithStack(err)
		}
		secret.StringData = data

		return nil
	})
	if err != nil {
		logger.Error(err, "failed to create tempo ingress auth secret")
		return ctrl.Result{}, errors.WithStack(err)
	}

	return ctrl.Result{}, nil
}
