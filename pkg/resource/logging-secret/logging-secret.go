package loggingsecret

import (
	"fmt"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/logging-operator/pkg/common"
)

const (
	loggingClientSecretName = "logging-secret"
)

func GenerateLoggingSecret(cluster *capi.Cluster, loggingAgent *common.LoggingAgent, loggingCredentialsSecret *v1.Secret, lokiURL string, installationName string, insecureCA bool) (v1.Secret, error) {
	var data map[string][]byte
	var err error

	switch loggingAgent.GetLoggingAgent() {
	case common.LoggingAgentPromtail:
		data, err = GeneratePromtailLoggingSecret(cluster, loggingCredentialsSecret, lokiURL, installationName, insecureCA)
		if err != nil {
			return v1.Secret{}, err
		}
	case common.LoggingAgentAlloy:
		data, err = GenerateAlloyLoggingSecret(cluster, loggingCredentialsSecret, lokiURL)
		if err != nil {
			return v1.Secret{}, err
		}
	default:
		return v1.Secret{}, errors.Errorf("unsupported logging agent %q", loggingAgent.GetLoggingAgent())
	}

	secret := v1.Secret{
		ObjectMeta: SecretMeta(cluster),
		Data:       data,
	}

	return secret, nil
}

// SecretMeta returns metadata for the logging-secret
func SecretMeta(cluster *capi.Cluster) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      GetLoggingSecretName(cluster),
		Namespace: cluster.GetNamespace(),
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

func GetLoggingSecretName(cluster *capi.Cluster) string {
	return fmt.Sprintf("%s-%s", cluster.GetName(), loggingClientSecretName)
}
