package eventsloggersecret

import (
	"fmt"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"
)

type values struct {
	ExtraSecret extraSecret `yaml:"extraSecret" json:"extraSecret"`
}

type extraSecret struct {
	Name string            `yaml:"name" json:"name"`
	Data map[string]string `yaml:"data" json:"data"`
}

// GenerateGrafanaAgentSecret returns a secret for
// the Loki-multi-tenant-proxy config
func generateGrafanaAgentSecret(lc loggedcluster.Interface, credentialsSecret *v1.Secret, lokiURL string) (map[string][]byte, error) {
	clusterName := lc.GetClusterName()
	writeUser := clusterName
	writePassword, err := loggingcredentials.GetPassword(lc, credentialsSecret, clusterName)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	values := values{
		ExtraSecret: extraSecret{
			Name: fmt.Sprintf("%s-%s", clusterName, common.GrafanaAgentExtraSecretName()),
			Data: map[string]string{
				common.LoggingURL:      fmt.Sprintf(common.LokiPushURLFormat, lokiURL),
				common.LoggingTenantID: common.DefaultWriteTenant,
				common.LoggingUsername: writeUser,
				common.LoggingPassword: writePassword,
			},
		},
	}

	v, err := yaml.Marshal(values)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	data := make(map[string][]byte)
	data["values"] = []byte(v)

	return data, nil
}
