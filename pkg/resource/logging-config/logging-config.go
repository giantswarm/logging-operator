package loggingconfig

import (
	"fmt"

	"github.com/blang/semver"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/logging-operator/pkg/common"
)

const (
	loggingConfigName = "logging-config"
)

func GenerateLoggingConfig(cluster *capi.Cluster, observabilityBundleVersion semver.Version, defaultNamespaces, tenants []string, installationName string, insecureCA bool) (v1.ConfigMap, error) {
	var values string
	var err error

	values, err = GenerateAlloyLoggingConfig(cluster, observabilityBundleVersion, defaultNamespaces, tenants, installationName, insecureCA)
	if err != nil {
		return v1.ConfigMap{}, err
	}

	configmap := v1.ConfigMap{
		ObjectMeta: ConfigMeta(cluster),
		Data: map[string]string{
			"values": values,
		},
	}

	return configmap, nil
}

// ConfigMeta returns metadata for the logging-config
func ConfigMeta(cluster *capi.Cluster) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      getLoggingConfigName(cluster),
		Namespace: cluster.GetNamespace(),
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

func getLoggingConfigName(cluster *capi.Cluster) string {
	return fmt.Sprintf("%s-%s", cluster.GetName(), loggingConfigName)
}
