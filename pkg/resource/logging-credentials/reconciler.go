package loggingcredentials

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/giantswarm/logging-operator/pkg/common"
)

// Resource implements a resource.Interface to handle
// Logging Credentials: store and maintain logging credentials
type Resource struct {
	Client client.Client
}

// ReconcileCreate ensures a secret exists for the given cluster.
func (r *Resource) ReconcileCreate(ctx context.Context, cluster *capi.Cluster, loggingAgent *common.LoggingAgent) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("loggingcredentials checking secret", "namespace", LoggingCredentialsSecretMeta().Namespace, "name", LoggingCredentialsSecretMeta().Name)

	// Start with some empty secret
	loggingCredentialsSecret := GenerateLoggingCredentialsBasicSecret(cluster)
	tracingCredentialsSecret := GenerateTracingCredentialsBasicSecret(cluster)

	// Retrieve existing secret if it exists
	err := r.Client.Get(ctx, types.NamespacedName{Name: LoggingCredentialsSecretMeta().Name, Namespace: LoggingCredentialsSecretMeta().Namespace}, loggingCredentialsSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("loggingcredentials secret not found, initializing one")
		} else {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	err = r.Client.Get(ctx, types.NamespacedName{Name: TracingCredentialsSecretMeta().Name, Namespace: TracingCredentialsSecretMeta().Namespace}, tracingCredentialsSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("tracing credentials secret not found, initializing one")
		} else {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	// update the secret's contents if needed
	loggingSecretUpdated, err := AddLoggingCredentials(cluster, loggingCredentialsSecret)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	tracingSecretUpdated, err := AddTracingCredentials(cluster, tracingCredentialsSecret)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Check if metadata has been updated
	if !reflect.DeepEqual(loggingCredentialsSecret.Labels, LoggingCredentialsSecretMeta().Labels) {
		logger.Info("loggingCredentials - metatada update required")
		loggingCredentialsSecret.ObjectMeta = LoggingCredentialsSecretMeta()
		loggingSecretUpdated = true
	}

	if !reflect.DeepEqual(tracingCredentialsSecret.Labels, TracingCredentialsSecretMeta().Labels) {
		logger.Info("tracingCredentials - metatada update required")
		tracingCredentialsSecret.ObjectMeta = TracingCredentialsSecretMeta()
		tracingSecretUpdated = true
	}

	if !loggingSecretUpdated {
		// If there were no changes, we're done here.
		logger.Info("loggingCredentials - up to date")
		return ctrl.Result{}, nil
	}

	if !tracingSecretUpdated {
		// If there were no changes, we're done here.
		logger.Info("tracingCredentials - up to date")
		return ctrl.Result{}, nil
	}

	// commit our changes
	logger.Info("loggingCredentials - Updating secret")
	err = r.Client.Update(ctx, loggingCredentialsSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("loggingCredentials - Secret does not exist, creating it")
			err = r.Client.Create(ctx, loggingCredentialsSecret)
			if err != nil {
				return ctrl.Result{}, errors.WithStack(err)
			}
		} else {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	logger.Info("tracingCredentials - Updating secret")
	err = r.Client.Update(ctx, tracingCredentialsSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("tracingCredentials - Secret does not exist, creating it")
			err = r.Client.Create(ctx, tracingCredentialsSecret)
		}
	}

	// Will return Secret's update error if any
	return ctrl.Result{}, errors.WithStack(err)
}

// ReconcileDelete ensures a secret is removed for the current cluster
func (r *Resource) ReconcileDelete(ctx context.Context, cluster *capi.Cluster, loggingAgent *common.LoggingAgent) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("loggingcredentials secret delete", "namespace", LoggingCredentialsSecretMeta().Namespace, "name", LoggingCredentialsSecretMeta().Name)

	// Start with some empty secret
	loggingCredentialsSecret := GenerateLoggingCredentialsBasicSecret(cluster)
	tracingCredentialsSecret := GenerateTracingCredentialsBasicSecret(cluster)

	// Retrieve existing secret
	err := r.Client.Get(ctx, types.NamespacedName{Name: LoggingCredentialsSecretMeta().Name, Namespace: LoggingCredentialsSecretMeta().Namespace}, loggingCredentialsSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("loggingcredentials secret not found, initializing one")
			return ctrl.Result{}, nil
		} else {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	err = r.Client.Get(ctx, types.NamespacedName{Name: TracingCredentialsSecretMeta().Name, Namespace: TracingCredentialsSecretMeta().Namespace}, tracingCredentialsSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("tracing credentials secret not found, initializing one")
		} else {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	// update the secret's contents if needed
	loggingSecretUpdated := RemoveLoggingCredentials(cluster, loggingCredentialsSecret)
	tracingSecretUpdated := RemoveTracingCredentials(cluster, tracingCredentialsSecret)

	// Check if metadata has been updated
	if !reflect.DeepEqual(loggingCredentialsSecret.Labels, LoggingCredentialsSecretMeta().Labels) {
		logger.Info("loggingCredentials - metatada update required")
		loggingCredentialsSecret.ObjectMeta = LoggingCredentialsSecretMeta()
		loggingSecretUpdated = true
	}

	if !reflect.DeepEqual(tracingCredentialsSecret.Labels, TracingCredentialsSecretMeta().Labels) {
		logger.Info("tracingCredentials - metatada update required")
		tracingCredentialsSecret.ObjectMeta = TracingCredentialsSecretMeta()
		tracingSecretUpdated = true
	}

	if !loggingSecretUpdated {
		// If there were no changes, we're done here.
		logger.Info("loggingCredentials - up to date")
		return ctrl.Result{}, nil
	}

	if !tracingSecretUpdated {
		// If there were no changes, we're done here.
		logger.Info("tracingCredentials - up to date")
		return ctrl.Result{}, nil
	}

	// commit our changes
	logger.Info("loggingCredentials - Updating secret")
	err = r.Client.Update(ctx, loggingCredentialsSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("loggingCredentials - Secret does not exist, creating it")
			err = r.Client.Create(ctx, loggingCredentialsSecret)
			if err != nil {
				return ctrl.Result{}, errors.WithStack(err)
			}
		} else {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	logger.Info("tracingCredentials - Updating secret")
	err = r.Client.Update(ctx, tracingCredentialsSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("tracingCredentials - Secret does not exist, creating it")
			err = r.Client.Create(ctx, tracingCredentialsSecret)
		}
	}

	// Will return Secret's update error if any
	return ctrl.Result{}, errors.WithStack(err)
}
