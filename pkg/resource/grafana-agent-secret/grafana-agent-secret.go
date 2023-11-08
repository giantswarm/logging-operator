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

type values struct {
	ExtraSecret extraSecret `yaml:"extraSecret" json:"extraSecret"`
}

type extraSecret struct {
	Name string            `yaml:"name" json:"name"`
	Data map[string]string `yaml:"data" json:"data"`
}

// SecretMeta returns metadata for the grafana-agent-secret
func SecretMeta(lc loggedcluster.Interface) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      fmt.Sprintf("%s-%s", lc.GetClusterName(), common.GrafanaAgentResourceName()),
		Namespace: lc.GetAppsNamespace(),
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

// GenerateGrafanaAgentSecret returns a secret for
// the Loki-multi-tenant-proxy auth config
func GenerateGrafanaAgentSecret(lc loggedcluster.Interface, credentialsSecret *v1.Secret, lokiURL string) (v1.Secret, error) {
	clusterName := lc.GetClusterName()
	writeUser := clusterName
	writePassword, err := loggingcredentials.GetPassword(lc, credentialsSecret, clusterName)
	if err != nil {
		return v1.Secret{}, errors.WithStack(err)
	}

	values := values{
		ExtraSecret: extraSecret{
			Name: fmt.Sprintf("%s-%s", clusterName, common.GrafanaAgentResourceName()),
			Data: map[string]string{
				"logging-url":       fmt.Sprintf("https://%s/loki/api/v1/push", lokiURL),
				"logging-tenant-id": clusterName,
				"logging-username":  writeUser,
				"logging-password":  writePassword,
			},
		},
	}

	v, err := yaml.Marshal(values)
	if err != nil {
		return v1.Secret{}, errors.WithStack(err)
	}

	secret := v1.Secret{
		ObjectMeta: SecretMeta(lc),
		Data: map[string][]byte{
			"values": []byte(v),
		},
	}

	return secret, nil
}
