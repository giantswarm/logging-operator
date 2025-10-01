package eventsloggersecret

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pkg/errors"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggingsecret "github.com/giantswarm/logging-operator/pkg/resource/logging-secret"
)

const (
	eventsLoggerSecretName = "events-logger-secret" // #nosec G101
	grafanaAgentSecretName = "grafana-agent-secret" // #nosec G101
)

func generateEventsLoggerSecret(cluster *capi.Cluster, loggingAgent *common.LoggingAgent, loggingCredentialsSecret *v1.Secret, lokiURL string, tracingEnabled bool, tracingCredentialsSecret *v1.Secret) (v1.Secret, error) {
	var data map[string][]byte
	var err error

	switch loggingAgent.GetKubeEventsLogger() {
	case common.EventsLoggerGrafanaAgent:
		data, err = generateGrafanaAgentSecret(cluster, loggingCredentialsSecret, lokiURL)
		if err != nil {
			return v1.Secret{}, err
		}
	case common.EventsLoggerAlloy:
		// In the case of Alloy being the events logger, we reuse the secret generation from the logging-secret package
		data, err = loggingsecret.GenerateAlloyLoggingSecret(cluster, loggingCredentialsSecret, lokiURL, tracingEnabled, tracingCredentialsSecret)
		if err != nil {
			return v1.Secret{}, err
		}
	default:
		return v1.Secret{}, errors.Errorf("unsupported logging agent %q", loggingAgent.GetKubeEventsLogger())
	}

	secret := v1.Secret{
		ObjectMeta: secretMeta(cluster, loggingAgent),
		Data:       data,
	}

	return secret, nil
}

// SecretMeta returns metadata for the events-logger-secret
func secretMeta(cluster *capi.Cluster, loggingAgent *common.LoggingAgent) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      GetEventsLoggerSecretName(cluster, loggingAgent),
		Namespace: cluster.GetNamespace(),
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

func GetEventsLoggerSecretName(cluster *capi.Cluster, loggingAgent *common.LoggingAgent) string {
	switch loggingAgent.GetKubeEventsLogger() {
	case common.EventsLoggerGrafanaAgent:
		return fmt.Sprintf("%s-%s", cluster.GetName(), grafanaAgentSecretName)
	default:
		return fmt.Sprintf("%s-%s", cluster.GetName(), eventsLoggerSecretName)
	}
}
