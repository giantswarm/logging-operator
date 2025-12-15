package eventsloggersecret

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	capi "sigs.k8s.io/cluster-api/api/core/v1beta1" //nolint:staticcheck // SA1019 deprecated package

	"github.com/giantswarm/logging-operator/pkg/common"
	loggingsecret "github.com/giantswarm/logging-operator/pkg/resource/logging-secret"
)

const (
	eventsLoggerSecretName = "events-logger-secret" // #nosec G101
)

func generateEventsLoggerSecret(cluster *capi.Cluster, loggingCredentialsSecret *v1.Secret, lokiURL string, tracingEnabled bool, tracingCredentialsSecret *v1.Secret) (v1.Secret, error) {
	var data map[string][]byte
	var err error

	// In the case of Alloy being the events logger, we reuse the secret generation from the logging-secret package
	data, err = loggingsecret.GenerateAlloyLoggingSecret(cluster, loggingCredentialsSecret, lokiURL, tracingEnabled, tracingCredentialsSecret)
	if err != nil {
		return v1.Secret{}, err
	}

	secret := v1.Secret{
		ObjectMeta: secretMeta(cluster),
		Data:       data,
	}

	return secret, nil
}

// SecretMeta returns metadata for the events-logger-secret
func secretMeta(cluster *capi.Cluster) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      getEventsLoggerSecretName(cluster),
		Namespace: cluster.GetNamespace(),
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

func getEventsLoggerSecretName(cluster *capi.Cluster) string {
	return fmt.Sprintf("%s-%s", cluster.GetName(), eventsLoggerSecretName)
}
