package loggingsecret

import (
	"fmt"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"
)

type values struct {
	Promtail promtail `yaml:"promtail" json:"promtail"`
}

type promtail struct {
	Config promtaiclusteronfig `yaml:"config" json:"config"`
}

type promtaiclusteronfig struct {
	Clients []promtaiclusteronfigClient `yaml:"clients" json:"clients"`
}

// TODO: use upstream promtail structures
type promtaiclusteronfigClient struct {
	URL            string                                  `yaml:"url" json:"url"`
	TenantID       string                                  `yaml:"tenant_id" json:"tenant_id"`
	BasicAuth      promtaiclusteronfigClientBasicAuth      `yaml:"basic_auth" json:"basic_auth"`
	BackoffConfig  promtaiclusteronfigClientBackoffConfig  `yaml:"backoff_config" json:"backoff_config"`
	ExternalLabels promtaiclusteronfigClientExternalLabels `yaml:"external_labels" json:"external_labels"`
	TLSConfig      promtaiclusteronfigClientTLSConfig      `yaml:"tls_config" json:"tls_config"`
	Timeout        string                                  `yaml:"timeout" json:"timeout"`
}

type promtaiclusteronfigClientTLSConfig struct {
	InsecureSkipVerify bool `yaml:"insecure_skip_verify" json:"insecure_skip_verify"`
}

type promtaiclusteronfigClientExternalLabels struct {
	Installation string `yaml:"installation" json:"installation"`
	ClusterID    string `yaml:"cluster_id" json:"cluster_id"`
}

type promtaiclusteronfigClientBackoffConfig struct {
	MaxPeriod string `yaml:"max_period" json:"max_period"`
}

type promtaiclusteronfigClientBasicAuth struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

// GeneratePromtailLoggingSecret returns a secret for
// the Loki-multi-tenant-proxy config
func GeneratePromtailLoggingSecret(cluster *capi.Cluster, credentialsSecret *v1.Secret, lokiURL string, installationName string, insecureCA bool) (map[string][]byte, error) {
	clusterName := cluster.GetName()

	writeUser := clusterName

	writePassword, err := loggingcredentials.GetPassword(cluster, credentialsSecret, clusterName)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	values := values{
		Promtail: promtail{
			Config: promtaiclusteronfig{
				Clients: []promtaiclusteronfigClient{
					{
						URL:      fmt.Sprintf(common.LokiPushURLFormat, lokiURL),
						TenantID: common.DefaultWriteTenant,
						Timeout:  common.LokiRemoteTimeout.String(),
						BasicAuth: promtaiclusteronfigClientBasicAuth{
							Username: writeUser,
							Password: writePassword,
						},
						BackoffConfig: promtaiclusteronfigClientBackoffConfig{
							MaxPeriod: common.LokiMaxBackoffPeriod.String(),
						},
						ExternalLabels: promtaiclusteronfigClientExternalLabels{
							Installation: installationName,
							ClusterID:    clusterName,
						},
						TLSConfig: promtaiclusteronfigClientTLSConfig{
							InsecureSkipVerify: insecureCA,
						},
					},
				},
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
