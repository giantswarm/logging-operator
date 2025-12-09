package credentials

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	capi "sigs.k8s.io/cluster-api/api/core/v1beta1"

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

// ReconcileCreate ensures a secret exists for the given cluster.
func (r *Resource) ReconcileCreate(ctx context.Context, cluster *capi.Cluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("credentials checking secret for logging", "namespace", CredentialsSecretMeta(LoggingCredentialsName, LoggingCredentialsNamespace).Namespace, "name", CredentialsSecretMeta(LoggingCredentialsName, LoggingCredentialsNamespace).Name)

	// Start with some empty secret
	loggingCredentialsSecret := GenerateLoggingCredentialsBasicSecret(cluster)
	_, err := r.createCredentialsSecret(ctx, cluster, loggingCredentialsSecret, LoggingCredentialsName, LoggingCredentialsNamespace)
	if err != nil {
		logger.Error(err, "failed to create or update logging credentials secret")
		return ctrl.Result{}, errors.WithStack(err)
	}

	if r.Config.EnableTracingFlag {
		logger.Info("credentials checking secret for tracing", "namespace", CredentialsSecretMeta(TracingCredentialsName, TracingCredentialsNamespace).Namespace, "name", CredentialsSecretMeta(TracingCredentialsName, TracingCredentialsNamespace).Name)

		tracingCredentialsSecret := GenerateTracingCredentialsBasicSecret(cluster)
		_, err = r.createCredentialsSecret(ctx, cluster, tracingCredentialsSecret, TracingCredentialsName, TracingCredentialsNamespace)
		if err != nil {
			logger.Error(err, "failed to create or update tracing credentials secret")
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	// Will return Secret's update error if any
	return ctrl.Result{}, errors.WithStack(err)
}

// ReconcileDelete ensures a secret is removed for the current cluster
func (r *Resource) ReconcileDelete(ctx context.Context, cluster *capi.Cluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("credentials secret delete logging", "namespace", CredentialsSecretMeta(LoggingCredentialsName, LoggingCredentialsNamespace).Namespace, "name", CredentialsSecretMeta(LoggingCredentialsName, LoggingCredentialsNamespace).Name)

	// Start with some empty secret
	loggingCredentialsSecret := GenerateLoggingCredentialsBasicSecret(cluster)

	_, err := r.deleteCredentialsSecret(ctx, cluster, loggingCredentialsSecret, LoggingCredentialsName, LoggingCredentialsNamespace)
	if err != nil {
		logger.Error(err, "failed to delete logging credentials secret")
		return ctrl.Result{}, errors.WithStack(err)
	}

	if r.Config.EnableTracingFlag {
		logger.Info("credentials secret delete for tracing", "namespace", CredentialsSecretMeta(TracingCredentialsName, TracingCredentialsNamespace).Namespace, "name", CredentialsSecretMeta(TracingCredentialsName, TracingCredentialsNamespace).Name)

		tracingCredentialsSecret := GenerateTracingCredentialsBasicSecret(cluster)
		_, err = r.deleteCredentialsSecret(ctx, cluster, tracingCredentialsSecret, TracingCredentialsName, TracingCredentialsNamespace)
		if err != nil {
			logger.Error(err, "failed to delete tracing credentials secret")
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	// Will return Secret's update error if any
	return ctrl.Result{}, errors.WithStack(err)
}

func (r *Resource) createCredentialsSecret(ctx context.Context, cluster *capi.Cluster, credentialsSecret *v1.Secret, secretName string, secretNamespace string) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Retrieve existing secret if it exists
	err := r.Client.Get(ctx, types.NamespacedName{Name: CredentialsSecretMeta(secretName, secretNamespace).Name, Namespace: CredentialsSecretMeta(secretName, secretNamespace).Namespace}, credentialsSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("loggingcredentials secret not found, initializing one")
		} else {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	// update the secret's contents if needed
	loggingSecretUpdated, err := AddCredentials(cluster, credentialsSecret)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Check if metadata has been updated
	if !reflect.DeepEqual(credentialsSecret.Labels, CredentialsSecretMeta(secretName, secretNamespace).Labels) {
		logger.Info("loggingCredentials - metatada update required")
		credentialsSecret.ObjectMeta = CredentialsSecretMeta(secretName, secretNamespace)
		loggingSecretUpdated = true
	}

	if !loggingSecretUpdated {
		// If there were no changes to either secret, we're done here.
		logger.Info("loggingCredentials and tracingCredentials - up to date")
		return ctrl.Result{}, nil
	}

	// commit our changes
	if loggingSecretUpdated {
		logger.Info("loggingCredentials - Updating secret")
		err = r.Client.Update(ctx, credentialsSecret)
		if err != nil {
			if apimachineryerrors.IsNotFound(err) {
				logger.Info("loggingCredentials - Secret does not exist, creating it")
				err = r.Client.Create(ctx, credentialsSecret)
				if err != nil {
					return ctrl.Result{}, errors.WithStack(err)
				}
			} else {
				return ctrl.Result{}, errors.WithStack(err)
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *Resource) deleteCredentialsSecret(ctx context.Context, cluster *capi.Cluster, credentialsSecret *v1.Secret, secretName string, secretNamespace string) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Retrieve existing secret
	err := r.Client.Get(ctx, types.NamespacedName{Name: CredentialsSecretMeta(secretName, secretNamespace).Name, Namespace: CredentialsSecretMeta(secretName, secretNamespace).Namespace}, credentialsSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("loggingcredentials secret not found, initializing one")
			return ctrl.Result{}, nil
		} else {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	// update the secret's contents if needed
	loggingSecretUpdated := RemoveCredentials(cluster, credentialsSecret)

	// Check if metadata has been updated
	if !reflect.DeepEqual(credentialsSecret.Labels, CredentialsSecretMeta(secretName, secretNamespace).Labels) {
		logger.Info("loggingCredentials - metatada update required")
		credentialsSecret.ObjectMeta = CredentialsSecretMeta(secretName, secretNamespace)
		loggingSecretUpdated = true
	}

	if !loggingSecretUpdated {
		// If there were no changes to either secret, we're done here.
		logger.Info("loggingCredentials and tracingCredentials - up to date")
		return ctrl.Result{}, nil
	}

	// commit our changes
	if loggingSecretUpdated {
		logger.Info("loggingCredentials - Updating secret")
		err = r.Client.Update(ctx, credentialsSecret)
		if err != nil {
			if apimachineryerrors.IsNotFound(err) {
				logger.Info("loggingCredentials - Secret does not exist, creating it")
				err = r.Client.Create(ctx, credentialsSecret)
				if err != nil {
					return ctrl.Result{}, errors.WithStack(err)
				}
			} else {
				return ctrl.Result{}, errors.WithStack(err)
			}
		}
	}

	return ctrl.Result{}, nil
}
