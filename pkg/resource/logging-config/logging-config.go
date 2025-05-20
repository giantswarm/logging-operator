package loggingconfig

import (
	"context"
	"fmt"
	"slices"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/blang/semver"
	"github.com/pkg/errors"

	"github.com/giantswarm/observability-operator/api/v1alpha1"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

const (
	loggingConfigName = "logging-config"
)

func (r *Reconciler) generateLoggingConfig(lc loggedcluster.Interface, observabilityBundleVersion semver.Version, defaultNamespaces, tenants []string) (v1.ConfigMap, error) {
	var values string
	var err error

	switch lc.GetLoggingAgent() {
	case common.LoggingAgentPromtail:
		values, err = r.generatePromtailLoggingConfig(lc)
		if err != nil {
			return v1.ConfigMap{}, err
		}
	case common.LoggingAgentAlloy:
		values, err = r.generateAlloyLoggingConfig(lc, observabilityBundleVersion, defaultNamespaces, tenants)
		if err != nil {
			return v1.ConfigMap{}, err
		}
	default:
		return v1.ConfigMap{}, errors.Errorf("unsupported logging agent %q", lc.GetLoggingAgent())
	}

	configmap := v1.ConfigMap{
		ObjectMeta: ConfigMeta(lc),
		Data: map[string]string{
			"values": values,
		},
	}

	return configmap, nil
}

// ConfigMeta returns metadata for the logging-config
func ConfigMeta(lc loggedcluster.Interface) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      getLoggingConfigName(lc),
		Namespace: lc.GetNamespace(),
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

func getLoggingConfigName(lc loggedcluster.Interface) string {
	return fmt.Sprintf("%s-%s", lc.GetClusterName(), loggingConfigName)
}

func listTenants(k8sClient client.Client, ctx context.Context) ([]string, error) {
	tenants := make([]string, 0)
	var grafanaOrganizations v1alpha1.GrafanaOrganizationList

	err := k8sClient.List(ctx, &grafanaOrganizations)
	if err != nil {
		return nil, err
	}

	for _, organization := range grafanaOrganizations.Items {
		if !organization.DeletionTimestamp.IsZero() {
			continue
		}

		for _, tenant := range organization.Spec.Tenants {
			if !slices.Contains(tenants, string(tenant)) {
				tenants = append(tenants, string(tenant))
			}
		}
	}

	return tenants, nil
}
