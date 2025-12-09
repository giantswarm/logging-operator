package common

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	capi "sigs.k8s.io/cluster-api/api/core/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/blang/semver"
	appv1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
)

const (
	observabilityBundleConfigMapName        = "observability-bundle-logging-extraconfig"
	observabilityBundleAppName       string = "observability-bundle"
)

// ObservabilityBundleAppMeta returns metadata for the observability bundle app.
func ObservabilityBundleAppMeta(cluster *capi.Cluster) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      AppConfigName(cluster, observabilityBundleAppName),
		Namespace: cluster.GetNamespace(),
		Labels:    map[string]string{},
	}

	AddCommonLabels(metadata.Labels)
	return metadata
}

// ObservabilityBundleConfigMapMeta returns metadata for the observability bundle extra values configmap.
func ObservabilityBundleConfigMapMeta(cluster *capi.Cluster) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      AppConfigName(cluster, observabilityBundleConfigMapName),
		Namespace: cluster.GetNamespace(),
		Labels: map[string]string{
			// This label is used by cluster-operator to find extraconfig. This only works on vintage WCs
			"app.kubernetes.io/name": observabilityBundleAppName,
		},
	}

	AddCommonLabels(metadata.Labels)
	return metadata
}

func GetObservabilityBundleAppVersion(ctx context.Context, client client.Client, cluster *capi.Cluster) (version semver.Version, err error) {
	// Get observability bundle app metadata.
	appMeta := ObservabilityBundleAppMeta(cluster)
	// Retrieve the app.
	var currentApp appv1.App
	err = client.Get(ctx, types.NamespacedName{Name: appMeta.GetName(), Namespace: appMeta.GetNamespace()}, &currentApp)
	if err != nil {
		return version, err
	}
	return semver.Parse(currentApp.Spec.Version)
}
