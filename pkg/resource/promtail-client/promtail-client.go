package promtailclient

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"
)

const (
	promtailClientSecretName = "logging-secret"
	lokiIngressNamespace     = "loki"
	lokiIngressName          = "loki-gateway"
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

// SecretMeta returns metadata for the promtail-user-secrets
func SecretMeta(lc loggedcluster.Interface) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      fmt.Sprintf("%s-%s", lc.GetClusterName(), promtailClientSecretName),
		Namespace: lc.GetAppsNamespace(),
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

// GeneratePromtailClientSecret returns a secret for
// the Loki-multi-tenant-proxy auth config
func GeneratePromtailClientSecret(lc loggedcluster.Interface, credentialsSecret *v1.Secret, lokiURL string) (v1.Secret, error) {

	clusterName := lc.GetClusterName()

	writeUser := clusterName

	writePassword, err := loggingcredentials.GetPassword(lc, credentialsSecret, clusterName)
	if err != nil {
		return v1.Secret{}, errors.WithStack(err)
	}

	installationName := lc.GetInstallationName()

	values := values{
		Promtail: promtail{
			Config: promtailConfig{
				Clients: []promtailConfigClient{
					{
						URL:      fmt.Sprintf("https://%s/loki/api/v1/push", lokiURL),
						TenantID: clusterName,
						BasicAuth: promtailConfigClientBasicAuth{
							Username: writeUser,
							Password: writePassword,
						},
						BackoffConfig: promtailConfigClientBackoffConfig{
							MaxPeriod: "10m",
						},
						ExternalLabels: promtailConfigClientExternalLabels{
							Installation: installationName,
							ClusterID:    lc.GetClusterName(),
						},
					},
				},
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

// Read Loki URL from ingress
func readLokiIngressURL(ctx context.Context, lc loggedcluster.Interface, client client.Client) (string, error) {

	var lokiIngress netv1.Ingress

	err := client.Get(ctx, types.NamespacedName{Name: lokiIngressName, Namespace: lokiIngressNamespace}, &lokiIngress)
	if err != nil {
		return "", errors.WithStack(err)
	}

	// We consider there's only one rule with one URL, because that's how the helm chart does it for the moment.
	ingressURL := lokiIngress.Spec.Rules[0].Host

	return ingressURL, nil
}

func GetPromtailCredentials(lc loggedcluster.Interface, credentialsSecret *v1.Secret) (map[string]string, error) {
	var promtailYaml promtail

	promtailSecret, ok := credentialsSecret.Data["values"]
	if !ok {
		return nil, errors.New("Promtail secret content not found")
	}

	err := yaml.Unmarshal(promtailSecret, &promtailYaml)
	if err != nil {
		return nil, errors.New("Invalid promtail secret content")
	}

	var credentials = make(map[string]string)
	credentials["url"] = promtailYaml.Config.Clients[0].URL
	credentials["tenantID"] = promtailYaml.Config.Clients[0].TenantID
	credentials["username"] = promtailYaml.Config.Clients[0].BasicAuth.Username
	credentials["installation"] = promtailYaml.Config.Clients[0].ExternalLabels.Installation
	credentials["clusterID"] = promtailYaml.Config.Clients[0].ExternalLabels.ClusterID

	return credentials, nil
}
