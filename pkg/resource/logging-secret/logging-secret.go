package loggingsecret

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/core/v1beta1" //nolint:staticcheck // SA1019 deprecated package

	"github.com/giantswarm/observability-operator/pkg/auth"

	"github.com/giantswarm/logging-operator/pkg/common"
)

const (
	loggingClientSecretName = "logging-secret"
)

func GenerateLoggingSecret(ctx context.Context, cluster *capi.Cluster, logsAuthManager auth.AuthManager, tracesAuthManager auth.AuthManager, lokiURL string, tracingEnabled bool) (v1.Secret, error) {
	var data map[string][]byte
	var err error

	data, err = GenerateAlloyLoggingSecret(ctx, cluster, logsAuthManager, tracesAuthManager, lokiURL, tracingEnabled)
	if err != nil {
		return v1.Secret{}, err
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
		Name:      getLoggingSecretName(cluster),
		Namespace: cluster.GetNamespace(),
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

func getLoggingSecretName(cluster *capi.Cluster) string {
	return fmt.Sprintf("%s-%s", cluster.GetName(), loggingClientSecretName)
}
