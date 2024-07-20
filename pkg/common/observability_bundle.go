package common

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/blang/semver"
	appv1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"

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

// GetObservabilityBundleAppVersion returns the version of the observability bundle app.
// It caches the version in the context to avoid multiple calls to the API.
func GetObservabilityBundleAppVersion(lc loggedcluster.Interface, client client.Client, ctx context.Context) (version semver.Version, err error) {
	version, ok := ctx.Value(observabilityBundleAppVersionKey).(semver.Version)
	if ok {
		return version, nil
	}

	// Get observability bundle app metadata.
	appMeta := ObservabilityBundleAppMeta(lc)
	// Retrieve the app.
	var currentApp appv1.App
	err = client.Get(ctx, types.NamespacedName{Name: appMeta.GetName(), Namespace: appMeta.GetNamespace()}, &currentApp)
	if err != nil {
		return version, err
	}

	version, err = semver.Parse(currentApp.Spec.Version)
	if err != nil {
		return version, err
	}

	ctx = context.WithValue(ctx, observabilityBundleAppVersionKey, version)

	return version, nil
}
