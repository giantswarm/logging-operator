package loggingsecret

import (
	"fmt"

	v1 "k8s.io/api/core/v1"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"
)

const (
	AlloyLokiURLEnvVarName           = "LOKI_URL"
	AlloyTenantIDEnvVarName          = "TENANT_ID"
	AlloyBasicAuthUsernameEnvVarName = "BASIC_AUTH_USERNAME"
	AlloyBasicAuthPasswordEnvVarName = "BASIC_AUTH_PASSWORD"
)

func GenerateAlloyLoggingSecret(lc loggedcluster.Interface, credentialsSecret *v1.Secret, lokiURL string) (map[string][]byte, error) {
	clusterName := lc.GetClusterName()

	writePassword, err := loggingcredentials.GetPassword(lc, credentialsSecret, clusterName)
	if err != nil {
		return nil, err
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

	return data, nil
}
