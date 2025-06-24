package eventsloggerconfig

import (
	"fmt"

        "github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

const (
	eventsLogggerConfigName = "events-logger-config"
	grafanaAgentConfigName  = "grafana-agent-config"
)

func generateEventsLoggerConfig(lc loggedcluster.Interface, includeNamespaces []string, excludeNamespaces []string, installationName string, insecureCA bool) (v1.ConfigMap, error) {
	var values string
	var err error

	switch lc.GetKubeEventsLogger() {
	case common.EventsLoggerGrafanaAgent:
		values, err = generateGrafanaAgentConfig(lc, includeNamespaces, excludeNamespaces, installationName, insecureCA)
		if err != nil {
			return v1.ConfigMap{}, err
		}
	case common.EventsLoggerAlloy:
		values, err = generateAlloyEventsConfig(lc, includeNamespaces, excludeNamespaces, installationName, insecureCA)
		if err != nil {
			return v1.ConfigMap{}, err
		}
	default:
		return v1.ConfigMap{}, errors.Errorf("unsupported events logger %q", lc.GetKubeEventsLogger())
	}

	configmap := v1.ConfigMap{
		ObjectMeta: configMeta(lc),
		Data: map[string]string{
			"values": values,
		},
	}

	return configmap, nil
}

// ConfigMeta returns metadata for the logging-config
func configMeta(lc loggedcluster.Interface) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      getEventsLoggerConfigName(lc),
		Namespace: lc.GetNamespace(),
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

func getEventsLoggerConfigName(lc loggedcluster.Interface) string {
	switch lc.GetKubeEventsLogger() {
	case common.EventsLoggerGrafanaAgent:
		return fmt.Sprintf("%s-%s", lc.GetName(), grafanaAgentConfigName)
	default:
		return fmt.Sprintf("%s-%s", lc.GetName(), eventsLogggerConfigName)
	}
}
