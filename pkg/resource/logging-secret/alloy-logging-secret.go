package loggingsecret

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"
)

func GenerateAlloyLoggingSecret(cluster *capi.Cluster, credentialsSecret *v1.Secret, lokiURL string) (map[string][]byte, error) {
	clusterName := cluster.GetName()

	writePassword, err := loggingcredentials.GetPassword(cluster, credentialsSecret, clusterName)
	if err != nil {
		return nil, err
	}

	templateData := common.AlloySecretTemplateData{
		ExtraSecretEnv: map[string]string{
			common.LoggingURL:      fmt.Sprintf(common.LokiPushURLFormat, lokiURL),
			common.LoggingTenantID: common.DefaultWriteTenant,
			common.LoggingUsername: clusterName,
			common.LoggingPassword: writePassword,
			common.LokiRulerAPIURL: fmt.Sprintf(common.LokiBaseURLFormat, lokiURL),
		},
	}

	values, err := common.GenerateAlloySecretValues(templateData)
	if err != nil {
		return nil, err
	}

	data := make(map[string][]byte)
	data["values"] = values

	return data, nil
}
