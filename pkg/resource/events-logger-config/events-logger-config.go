package eventsloggerconfig

import (
	"fmt"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/logging-operator/pkg/common"
)

const (
	eventsLogggerConfigName = "events-logger-config"
	grafanaAgentConfigName  = "grafana-agent-config"
)

func generateEventsLoggerConfig(cluster *capi.Cluster, loggingAgent *common.LoggingAgent, includeNamespaces []string, excludeNamespaces []string, installationName string, insecureCA bool, tempoURL string) (v1.ConfigMap, error) {
	var values string
	var err error

	switch loggingAgent.GetKubeEventsLogger() {
	case common.EventsLoggerGrafanaAgent:
		values, err = generateGrafanaAgentConfig(cluster, loggingAgent, includeNamespaces, excludeNamespaces, installationName, insecureCA)
		if err != nil {
			return v1.ConfigMap{}, err
		}
	case common.EventsLoggerAlloy:
		values, err = generateAlloyEventsConfig(cluster, includeNamespaces, excludeNamespaces, installationName, insecureCA, tempoURL)
		if err != nil {
			return v1.ConfigMap{}, err
		}
	default:
		return v1.ConfigMap{}, errors.Errorf("unsupported events logger %q", loggingAgent.GetKubeEventsLogger())
	}

	configmap := v1.ConfigMap{
		ObjectMeta: configMeta(cluster, loggingAgent),
		Data: map[string]string{
			"values": values,
		},
	}

	return configmap, nil
}

// ConfigMeta returns metadata for the logging-config
func configMeta(cluster *capi.Cluster, loggingAgent *common.LoggingAgent) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      getEventsLoggerConfigName(cluster, loggingAgent),
		Namespace: cluster.GetNamespace(),
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

func getEventsLoggerConfigName(cluster *capi.Cluster, loggingAgent *common.LoggingAgent) string {
	switch loggingAgent.GetKubeEventsLogger() {
	case common.EventsLoggerGrafanaAgent:
		return fmt.Sprintf("%s-%s", cluster.GetName(), grafanaAgentConfigName)
	default:
		return fmt.Sprintf("%s-%s", cluster.GetName(), eventsLogggerConfigName)
	}
}
