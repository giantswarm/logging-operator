package eventsloggerconfig

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/logging-operator/pkg/common"
)

const (
	eventsLogggerConfigName = "events-logger-config"
)

func generateEventsLoggerConfig(cluster *capi.Cluster, tenants []string, includeNamespaces []string, excludeNamespaces []string, insecureCA bool, tracingEnabled bool, tempoURL string, clusterLabels common.ClusterLabels) (v1.ConfigMap, error) {
	var values string
	var err error

	values, err = generateAlloyEventsConfig(includeNamespaces, excludeNamespaces, insecureCA, tracingEnabled, tempoURL, tenants, clusterLabels)
	if err != nil {
		return v1.ConfigMap{}, err
	}

	configmap := v1.ConfigMap{
		ObjectMeta: configMeta(cluster),
		Data: map[string]string{
			"values": values,
		},
	}

	return configmap, nil
}

// ConfigMeta returns metadata for the logging-config
func configMeta(cluster *capi.Cluster) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      getEventsLoggerConfigName(cluster),
		Namespace: cluster.GetNamespace(),
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

func getEventsLoggerConfigName(cluster *capi.Cluster) string {
	return fmt.Sprintf("%s-%s", cluster.GetName(), eventsLogggerConfigName)
}
