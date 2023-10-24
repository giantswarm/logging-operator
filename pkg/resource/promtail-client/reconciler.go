package promtailclient

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Reconciler implements a reconciler.Interface to handle
// Promtail client: extra promtail config about where and how to send logs
type Reconciler struct {
	client.Client
}

// ReconcileCreate ensures promtail secret is created with the right credentials
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("promtailclient create")

	// Retrieve secret containing credentials
	var loggingCredentialsSecret v1.Secret
	err := r.Client.Get(ctx, types.NamespacedName{Name: loggingcredentials.LoggingCredentialsSecretMeta(lc).Name, Namespace: loggingcredentials.LoggingCredentialsSecretMeta(lc).Namespace},
		&loggingCredentialsSecret)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Retrieve Loki ingress name
	lokiURL, err := common.ReadLokiIngressURL(ctx, lc, r.Client)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Get desired secret
	desiredPromtailClientSecret, err := GeneratePromtailClientSecret(lc, &loggingCredentialsSecret, lokiURL)
	if err != nil {
		logger.Info("promtailclient - failed generating auth config!", "error", err)
		return ctrl.Result{}, errors.WithStack(err)
	}

	// Check if secret already exists.
	logger.Info("promtailclient - getting", "namespace", desiredPromtailClientSecret.GetNamespace(), "name", desiredPromtailClientSecret.GetName())
	var currentPromtailClientSecret v1.Secret
	err = r.Client.Get(ctx, types.NamespacedName{Name: desiredPromtailClientSecret.GetName(), Namespace: desiredPromtailClientSecret.GetNamespace()}, &currentPromtailClientSecret)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("promtailclient not found, creating")
			err = r.Client.Create(ctx, &desiredPromtailClientSecret)
			if err != nil {
				return ctrl.Result{}, errors.WithStack(err)
			}
		} else {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	if !needUpdate(currentPromtailClientSecret, desiredPromtailClientSecret) {
		logger.Info("promtailclient up to date")
		return ctrl.Result{}, nil
	}

	logger.Info("promtailclient - updating")
	err = r.Client.Update(ctx, &desiredPromtailClientSecret)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info("promtailclient - done")
	return ctrl.Result{}, nil
}

// ReconcileDelete - Not much to do here when a cluster is deleted
func (r *Reconciler) ReconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("promtailclient delete")

	return ctrl.Result{}, nil
}

// needUpdate return true if current.Data and desired.Data do not match.
func needUpdate(current, desired v1.Secret) bool {
	return !reflect.DeepEqual(current.Data, desired.Data)
}
