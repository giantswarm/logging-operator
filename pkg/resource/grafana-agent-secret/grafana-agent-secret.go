package grafanaagentsecret

import (
	"fmt"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"
)

const (
	grafanaAgentSecretName = "grafana-agent-secret"
)

type values struct {
	URL      string `yaml:"LOGGING_URL" json:"LOGGING_URL"`
	TenantID string `yaml:"LOGGING_TENANT_ID" json:"LOGGING_TENANT_ID"`
	Username string `yaml:"LOGGING_USERNAME" json:"LOGGING_USERNAME"`
	Password string `yaml:"LOGGING_PASSWORD" json:"LOGGING_PASSWORD"`
}

// SecretMeta returns metadata for the grafana-agent-secret
func SecretMeta(lc loggedcluster.Interface, secretNamespace string) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      fmt.Sprintf("%s-%s", lc.GetClusterName(), grafanaAgentSecretName),
		Namespace: secretNamespace,
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

// GenerateGrafanaAgentSecret returns a secret for
// the Loki-multi-tenant-proxy auth config
func GenerateGrafanaAgentSecret(lc loggedcluster.Interface, credentialsSecret *v1.Secret, lokiURL string, secretNamespace string) (v1.Secret, error) {

	clusterName := lc.GetClusterName()
	writeUser := clusterName
	writePassword, err := loggingcredentials.GetPassword(lc, credentialsSecret, clusterName)
	if err != nil {
		return v1.Secret{}, errors.WithStack(err)
	}

	values := values{
		URL:      fmt.Sprintf("https://%s/loki/api/v1/push", lokiURL),
		TenantID: clusterName,
		Username: writeUser,
		Password: writePassword,
	}

	v, err := yaml.Marshal(values)
	if err != nil {
		return v1.Secret{}, errors.WithStack(err)
	}

	secret := v1.Secret{
		ObjectMeta: SecretMeta(lc, secretNamespace),
		Data: map[string][]byte{
			"values": []byte(v),
		},
	}

	return secret, nil
}
