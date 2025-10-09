package loggingconfig

import (
	"fmt"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/logging-operator/pkg/common"
)

const (
	loggingConfigName = "logging-config"
)

func (r *Resource) GenerateLoggingConfig(cluster *capi.Cluster, loggingAgent *common.LoggingAgent, observabilityBundleVersion semver.Version, tenants []string, clusterLabels common.ClusterLabels) (v1.ConfigMap, error) {
	var values string
	var err error

	switch loggingAgent.LoggingAgent {
	case common.LoggingAgentPromtail:
		values, err = GeneratePromtailLoggingConfig(cluster, r.Config.InstallationName)
		if err != nil {
			return v1.ConfigMap{}, err
		}
	case common.LoggingAgentAlloy:
		values, err = GenerateAlloyLoggingConfig(loggingAgent, observabilityBundleVersion, r.DefaultWorkloadClusterNamespaces, tenants, clusterLabels, r.Config.InsecureCA)
		if err != nil {
			return v1.ConfigMap{}, err
		}
	default:
		return v1.ConfigMap{}, errors.Errorf("unsupported logging agent %q", loggingAgent.LoggingAgent)
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
