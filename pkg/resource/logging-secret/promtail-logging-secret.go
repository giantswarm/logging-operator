package loggingsecret

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
	Promtail promtail `yaml:"promtail" json:"promtail"`
}

type promtail struct {
	Config promtailConfig `yaml:"config" json:"config"`
}

type promtailConfig struct {
	Clients []promtailConfigClient `yaml:"clients" json:"clients"`
}

// TODO: use upstream promtail structures
type promtailConfigClient struct {
	URL            string                             `yaml:"url" json:"url"`
	TenantID       string                             `yaml:"tenant_id" json:"tenant_id"`
	BasicAuth      promtailConfigClientBasicAuth      `yaml:"basic_auth" json:"basic_auth"`
	BackoffConfig  promtailConfigClientBackoffConfig  `yaml:"backoff_config" json:"backoff_config"`
	ExternalLabels promtailConfigClientExternalLabels `yaml:"external_labels" json:"external_labels"`
	TLSConfig      promtailConfigClientTLSConfig      `yaml:"tls_config" json:"tls_config"`
	Timeout        string                             `yaml:"timeout" json:"timeout"`
}

type promtailConfigClientTLSConfig struct {
	InsecureSkipVerify bool `yaml:"insecure_skip_verify" json:"insecure_skip_verify"`
}

type promtailConfigClientExternalLabels struct {
	Installation string `yaml:"installation" json:"installation"`
	ClusterID    string `yaml:"cluster_id" json:"cluster_id"`
}

type promtailConfigClientBackoffConfig struct {
	MaxPeriod string `yaml:"max_period" json:"max_period"`
}

type promtailConfigClientBasicAuth struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

// GeneratePromtailLoggingSecret returns a secret for
// the Loki-multi-tenant-proxy config
func GeneratePromtailLoggingSecret(lc loggedcluster.Interface, credentialsSecret *v1.Secret, lokiURL string, installationName string, insecureCA bool) (map[string][]byte, error) {
	clusterName := lc.GetName()

	writeUser := clusterName

	writePassword, err := loggingcredentials.GetPassword(lc, credentialsSecret, clusterName)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	values := values{
		Promtail: promtail{
			Config: promtailConfig{
				Clients: []promtailConfigClient{
					{
						URL:      fmt.Sprintf(common.LokiPushURLFormat, lokiURL),
						TenantID: common.DefaultWriteTenant,
						Timeout:  common.LokiRemoteTimeout.String(),
						BasicAuth: promtailConfigClientBasicAuth{
							Username: writeUser,
							Password: writePassword,
						},
						BackoffConfig: promtailConfigClientBackoffConfig{
							MaxPeriod: common.LokiMaxBackoffPeriod.String(),
						},
						ExternalLabels: promtailConfigClientExternalLabels{
							Installation: installationName,
							ClusterID:    lc.GetName(),
						},
						TLSConfig: promtailConfigClientTLSConfig{
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
