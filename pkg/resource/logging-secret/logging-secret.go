package loggingsecret

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pkg/errors"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

const (
	loggingClientSecretName = "logging-secret"
)

func GenerateLoggingSecret(lc loggedcluster.Interface, loggingCredentialsSecret *v1.Secret, lokiURL string) (v1.Secret, error) {
	var data map[string][]byte
	var err error

	switch lc.GetLoggingAgent() {
	case "promtail":
		data, err = GeneratePromtailLoggingSecret(lc, loggingCredentialsSecret, lokiURL)
		if err != nil {
			return v1.Secret{}, err
		}
	case common.AlloyLogAgentName:
		data, err = GenerateAlloyLoggingSecret(lc, loggingCredentialsSecret, lokiURL)
		if err != nil {
			return v1.Secret{}, err
		}
	default:
		return v1.Secret{}, errors.Errorf("unsupported logging agent %q", lc.GetLoggingAgent())
	}

	secret := v1.Secret{
		ObjectMeta: SecretMeta(lc),
		Data:       data,
	}

	return secret, nil
}

// SecretMeta returns metadata for the logging-secret
func SecretMeta(lc loggedcluster.Interface) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      GetLoggingSecretName(lc),
		Namespace: lc.GetAppsNamespace(),
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

func GetLoggingSecretName(lc loggedcluster.Interface) string {
	return fmt.Sprintf("%s-%s", lc.GetClusterName(), loggingClientSecretName)
}
