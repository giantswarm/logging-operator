package tracingsecret

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"
)

// Resource implements a resource.Interface to handle
// Events-logger secret: extra events-logger secret about where and how to send logs (in this case : k8S events)
type Resource struct {
	Client client.Client
}

// ReconcileCreate ensures tracing secret is created with the right credentials
func (r *Resource) ReconcileCreate(ctx context.Context, cluster *capi.Cluster, loggingAgent *common.LoggingAgent) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("create tracing secret")

	// Retrieve secret containing credentials
	var eventsLoggerCredentialsSecret v1.Secret
	err := r.Client.Get(ctx, types.NamespacedName{Name: loggingcredentials.LoggingCredentialsSecretMeta().Name, Namespace: loggingcredentials.LoggingCredentialsSecretMeta().Namespace},
		&eventsLoggerCredentialsSecret)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Get desired secret
	desiredTracingSecret, err := generateTracingSecret(cluster)
	if err != nil {
		logger.Error(err, "failed generating tracing secret")
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Check if secret already exists.
	logger.Info("getting tracing secret", "namespace", desiredTracingSecret.GetNamespace(), "name", desiredTracingSecret.GetName())
	var currentEventsLoggerSecret v1.Secret
	err = r.Client.Get(ctx, types.NamespacedName{Name: desiredTracingSecret.GetName(), Namespace: desiredTracingSecret.GetNamespace()}, &currentEventsLoggerSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("tracing secret not found, creating")
			err = r.Client.Create(ctx, &desiredTracingSecret)
			if err != nil {
				return ctrl.Result{}, errors.WithStack(err)
			}
		} else {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	if !needUpdate(currentEventsLoggerSecret, desiredTracingSecret) {
		logger.Info("tracing secret up to date")
		return ctrl.Result{}, nil
	}

	logger.Info("updating tracing secret")
	err = r.Client.Update(ctx, &desiredTracingSecret)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info("updated tracing secret")
	return ctrl.Result{}, nil
}

// ReconcileDelete - Not much to do here when a cluster is deleted
func (r *Resource) ReconcileDelete(ctx context.Context, cluster *capi.Cluster, loggingAgent *common.LoggingAgent) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("delete tracing secret")

	// Get expected secret.
	var currentEventsLoggerSecret v1.Secret
	err := r.Client.Get(ctx, types.NamespacedName{Name: GetTracingSecretName(cluster), Namespace: cluster.GetNamespace()}, &currentEventsLoggerSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("tracing secret not found, stop here")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Delete secret.
	logger.Info("tracing secret deleting", "namespace", currentEventsLoggerSecret.GetNamespace(), "name", currentEventsLoggerSecret.GetName())
	err = r.Client.Delete(ctx, &currentEventsLoggerSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			// Do no throw error in case it was not found, as this means
			// it was already deleted.
			logger.Info("tracing secret already deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}
	logger.Info("tracing secret deleted")

	return ctrl.Result{}, nil
}

// needUpdate return true if current.Data and desired.Data do not match.
func needUpdate(current, desired v1.Secret) bool {
	return !reflect.DeepEqual(current.Data, desired.Data)
}
