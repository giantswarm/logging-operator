package loggingconfig

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pkg/errors"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

const (
	loggingConfigName = "logging-config"
)

func GenerateLoggingConfig(lc loggedcluster.Interface) (v1.ConfigMap, error) {
	var values string
	var err error

	switch lc.GetLoggingAgent() {
	case "promtail":
		values, err = GeneratePromtailLoggingConfig(lc)
		if err != nil {
			return v1.ConfigMap{}, err
		}
	case common.AlloyLogAgentAppName:
		values, err = GenerateAlloyLoggingConfig(lc)
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
