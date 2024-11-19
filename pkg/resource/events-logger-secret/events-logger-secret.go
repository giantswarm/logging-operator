package eventsloggersecret

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pkg/errors"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	loggingsecret "github.com/giantswarm/logging-operator/pkg/resource/logging-secret"
)

const (
	eventsLoggerSecretName = "events-logger-secret" // #nosec G101
	grafanaAgentSecretName = "grafana-agent-secret" // #nosec G101
)

func generateEventsLoggerSecret(lc loggedcluster.Interface, loggingCredentialsSecret *v1.Secret, lokiURL string) (v1.Secret, error) {
	var data map[string][]byte
	var err error

	switch lc.GetKubeEventsLogger() {
	case common.EventsLoggerGrafanaAgent:
		data, err = generateGrafanaAgentSecret(lc, loggingCredentialsSecret, lokiURL)
		if err != nil {
			return v1.Secret{}, err
		}
	case common.EventsLoggerAlloy:
		// In the case of Alloy being the events logger, we reuse the secret generation from the logging-secret package
		data, err = loggingsecret.GenerateAlloyLoggingSecret(lc, loggingCredentialsSecret, lokiURL)
		if err != nil {
			return v1.Secret{}, err
		}
	default:
		return v1.Secret{}, errors.Errorf("unsupported logging agent %q", lc.GetLoggingAgent())
	}

	secret := v1.Secret{
		ObjectMeta: secretMeta(lc),
		Data:       data,
	}

	return secret, nil
}

// SecretMeta returns metadata for the events-logger-secret
func secretMeta(lc loggedcluster.Interface) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      GetEventsLoggerSecretName(lc),
		Namespace: lc.GetAppsNamespace(),
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

func GetEventsLoggerSecretName(lc loggedcluster.Interface) string {
	switch lc.GetKubeEventsLogger() {
	case common.EventsLoggerGrafanaAgent:
		return fmt.Sprintf("%s-%s", lc.GetClusterName(), grafanaAgentSecretName)
	default:
		return fmt.Sprintf("%s-%s", lc.GetClusterName(), eventsLoggerSecretName)
	}
}
