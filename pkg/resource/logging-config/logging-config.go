package loggingconfig

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/blang/semver"
	"github.com/pkg/errors"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	"github.com/giantswarm/observability-operator/api/v1alpha1"
)

const (
	loggingConfigName = "logging-config"
)

func GenerateLoggingConfig(lc loggedcluster.Interface, observabilityBundleVersion semver.Version, defaultNamespaces, tenants []string) (v1.ConfigMap, error) {
	var values string
	var err error

	switch lc.GetLoggingAgent() {
	case common.LoggingAgentPromtail:
		values, err = GeneratePromtailLoggingConfig(lc, observabilityBundleVersion)
		if err != nil {
			return v1.ConfigMap{}, err
		}
	case common.LoggingAgentAlloy:
		values, err = GenerateAlloyLoggingConfig(lc, observabilityBundleVersion, defaultNamespaces, tenants)
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
		Namespace: lc.GetAppsNamespace(),
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

func getLoggingConfigName(lc loggedcluster.Interface) string {
	return fmt.Sprintf("%s-%s", lc.GetClusterName(), loggingConfigName)
}

func listTenants(k8sClient client.Client, ctx context.Context) ([]string, error) {
	tenantList := make([]string, 0)
	var grafaOrgList v1alpha1.GrafanaOrganizationList

	err := k8sClient.List(ctx, &grafaOrgList)
	if err != nil {
		return nil, err
	}

	for _, grafanaOrg := range grafaOrgList.Items {
		for _, tenant := range grafanaOrg.Spec.Tenants {
			tenantList = append(tenantList, string(tenant))
		}
	}

	return tenantList, nil
}
