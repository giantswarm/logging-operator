package k8seventsconfig

import (
	"fmt"

	"github.com/blang/semver"
	"github.com/pkg/errors"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

const eventsLogggerConfigName = "events-logger-config"

func GenerateEventsLoggerConfig(lc loggedcluster.Interface, observabilityBundleVersion semver.Version, defaultNamespaces []string) (v1.ConfigMap, error) {
	var values string
	var err error

	switch lc.GetKubeEventsLogger() {
	case common.EventsLoggerGrafanaAgent:
		values, err = GenerateGrafanaAgentConfig(lc, defaultNamespaces)
		if err != nil {
			return v1.ConfigMap{}, err
		}
	case common.EventsLoggerAlloy:
		values, err = GenerateAlloyEventsConfig(lc, observabilityBundleVersion, defaultNamespaces)
		if err != nil {
			return v1.ConfigMap{}, err
		}
	default:
		return v1.ConfigMap{}, errors.Errorf("unsupported events logger %q", lc.GetKubeEventsLogger())
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
		Name:      getEventsLoggerConfigName(lc),
		Namespace: lc.GetAppsNamespace(),
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

func getEventsLoggerConfigName(lc loggedcluster.Interface) string {
	return fmt.Sprintf("%s-%s", lc.GetClusterName(), eventsLogggerConfigName)
}
