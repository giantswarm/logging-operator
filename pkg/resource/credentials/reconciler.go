package credentials

import (
	"context"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/core/v1beta1" //nolint:staticcheck // SA1019 deprecated package

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/giantswarm/logging-operator/pkg/config"
)

// Resource implements a resource.Interface to handle
// Logging Credentials: store and maintain logging credentials
type Resource struct {
	Client client.Client
	Config config.Config
}

// ReconcileCreate always calls ReconcileDelete to clean up old credentials.
// Credential generation is now handled by observability-operator's auth manager.
func (r *Resource) ReconcileCreate(ctx context.Context, cluster *capi.Cluster) (ctrl.Result, error) {
	return r.ReconcileDelete(ctx, cluster)
}

// ReconcileDelete ensures a secret is removed for the current cluster
func (r *Resource) ReconcileDelete(ctx context.Context, cluster *capi.Cluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("credentials secret delete logging", "namespace", LoggingCredentialsNamespace, "name", LoggingCredentialsName)
	_, err := r.deleteSecret(ctx, LoggingCredentialsName, LoggingCredentialsNamespace)
	if err != nil {
		logger.Error(err, "failed to delete logging credentials secret")
		return ctrl.Result{}, errors.WithStack(err)
	}

	if r.Config.EnableTracingFlag {
		logger.Info("credentials secret delete for tracing", "namespace", TracingCredentialsNamespace, "name", TracingCredentialsName)

		_, err = r.deleteSecret(ctx, TracingCredentialsName, TracingCredentialsNamespace)
		if err != nil {
			logger.Error(err, "failed to delete tracing credentials secret")
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	// Will return Secret's update error if any
	return ctrl.Result{}, errors.WithStack(err)
}

func (r *Resource) deleteSecret(ctx context.Context, secretName string, secretNamespace string) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: secretNamespace,
		},
	}

	err := r.Client.Delete(ctx, &secret)
	if client.IgnoreNotFound(err) != nil {
		logger.Error(err, "failed to delete ingress auth secret")
		return ctrl.Result{}, errors.WithStack(err)
	}

	return ctrl.Result{}, nil
}
