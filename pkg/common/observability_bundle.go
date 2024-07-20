package common

import (
	"context"
	"time"

	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/blang/semver"
	appv1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/pkg/errors"

	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

const ObservabilityBundleAppName string = "observability-bundle"

// ObservabilityBundleAppMeta returns metadata for the observability bundle app.
func ObservabilityBundleAppMeta(lc loggedcluster.Interface) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      lc.AppConfigName(ObservabilityBundleAppName),
		Namespace: lc.GetAppsNamespace(),
		Labels:    map[string]string{},
	}

	AddCommonLabels(metadata.Labels)
	return metadata
}

// ObservabilityBundleConfigMapMeta returns metadata for the observability bundle extra values configmap.
func ObservabilityBundleConfigMapMeta(lc loggedcluster.Interface) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      lc.AppConfigName(lc.GetObservabilityBundleConfigMap()),
		Namespace: lc.GetAppsNamespace(),
		Labels: map[string]string{
			// This label is used by cluster-operator to find extraconfig. This only works on vintage WCs
			"app.kubernetes.io/name": lc.ObservabilityBundleConfigLabelName(ObservabilityBundleAppName),
		},
	}

	AddCommonLabels(metadata.Labels)
	return metadata
}

type observabilityBundleAppVersionContextKey int

var observabilityBundleAppVersionKey observabilityBundleAppVersionContextKey

// NewObservabilityBundleAppVersionContext retrieves the observability bundle app version from the API and stores it in the context.
func NewObservabilityBundleAppVersionContext(ctx context.Context, lc loggedcluster.Interface, client client.Client) (context.Context, ctrl.Result, error) {
	// Get observability bundle app metadata.
	appMeta := ObservabilityBundleAppMeta(lc)
	// Retrieve the app.
	var currentApp appv1.App
	err := client.Get(ctx, types.NamespacedName{Name: appMeta.GetName(), Namespace: appMeta.GetNamespace()}, &currentApp)
	if err != nil {
		// Handle case where the app is not found.
		if apimachineryerrors.IsNotFound(err) {
			// If the app is not found we should requeue and try again later (5 minutes is the app platform default reconciliation time)
			return nil, ctrl.Result{RequeueAfter: time.Duration(5 * time.Minute)}, nil
		}
		return nil, ctrl.Result{}, errors.WithStack(err)
	}

	version, err := semver.Parse(currentApp.Spec.Version)
	if err != nil {
		return nil, ctrl.Result{}, err
	}

	ctx = context.WithValue(ctx, observabilityBundleAppVersionKey, version)

	return ctx, ctrl.Result{}, nil
}

func ObservabilityBundleAppVersionFromContext(ctx context.Context) (semver.Version, bool) {
	version, ok := ctx.Value(observabilityBundleAppVersionKey).(semver.Version)
	return version, ok
}
