package alloysecret

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"
)

const (
	SecretName = "alloy-secret"

	AlloyLokiURLEnvVarName           = "LOKI_URL"
	AlloyTenantIDEnvVarName          = "TENANT_ID"
	AlloyBasicAuthUsernameEnvVarName = "BASIC_AUTH_USERNAME"
	AlloyBasicAuthPasswordEnvVarName = "BASIC_AUTH_PASSWORD" // #nosec G101
)

func GenerateAlloyLoggingSecret(lc loggedcluster.Interface, credentialsSecret *v1.Secret, lokiURL string) (v1.Secret, error) {
	clusterName := lc.GetClusterName()

	writePassword, err := loggingcredentials.GetPassword(lc, credentialsSecret, clusterName)
	if err != nil {
		return v1.Secret{}, err
	}

	LokiURL := fmt.Sprintf(common.LokiURLFormat, lokiURL)
	TenantID := clusterName
	BasicAuthUsername := clusterName
	BasicAuthPassword := writePassword

	data := make(map[string][]byte)
	data[AlloyLokiURLEnvVarName] = []byte(LokiURL)
	data[AlloyTenantIDEnvVarName] = []byte(TenantID)
	data[AlloyBasicAuthUsernameEnvVarName] = []byte(BasicAuthUsername)
	data[AlloyBasicAuthPasswordEnvVarName] = []byte(BasicAuthPassword)

	secret := v1.Secret{
		ObjectMeta: SecretMeta(lc),
		Data:       data,
	}

	return secret, nil
}

// SecretMeta returns metadata for the Alloy secret.
func SecretMeta(lc loggedcluster.Interface) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      SecretName,
		Namespace: common.AlloyLogAgentAppNamespace,
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}
