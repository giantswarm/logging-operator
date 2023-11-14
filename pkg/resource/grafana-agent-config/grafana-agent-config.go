package grafanaagentconfig

import (
	"fmt"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

const (
	grafanaAgentConfigName = "grafana-agent-config"
)

// /// Grafana-Agent values config structure
type values struct {
	GrafanaAgent grafanaAgent `yaml:"grafana-agent" json:"grafana-agent"`
}

type grafanaAgent struct {
	Agent agent `yaml:"agent" json:"agent"`
}

type agent struct {
	ConfigMap configMap `yaml:"configMap" json:"configMap"`
}

type configMap struct {
	Content string `yaml:"content" json:"content"`
}

// ConfigMeta returns metadata for the grafana-agent-config
func ConfigMeta(lc loggedcluster.Interface) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      fmt.Sprintf("%s-%s", lc.GetClusterName(), grafanaAgentConfigName),
		Namespace: lc.GetAppsNamespace(),
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

// GenerateGrafanaAgentConfig returns a configmap for
// the grafana-agent extra-config
func GenerateGrafanaAgentConfig(lc loggedcluster.Interface, appNamespace string) (v1.ConfigMap, error) {
	scrapedNamespaces := "[]"
	if common.IsWorkloadCluster(lc) {
		scrapedNamespaces = "[\"kube-system\", \"giantswarm\"]"
	}

	values := values{
		GrafanaAgent: grafanaAgent{
			Agent: agent{
				ConfigMap: configMap{
					Content: `
logging {
	level  = "info"
	format = "logfmt"
}

loki.source.kubernetes_events "local" {
	namespaces = ` + scrapedNamespaces + `
	forward_to = [loki.write.default.receiver]
}

remote.kubernetes.secret "credentials" {
	namespace = "` + appNamespace + `"
	name = "` + fmt.Sprintf("%s-%s", lc.GetClusterName(), common.GrafanaAgentExtraSecretName()) + `"
}

loki.write "default" {
	endpoint {
	url = nonsensitive(remote.kubernetes.secret.credentials.data["logging-url"])
	tenant_id = nonsensitive(remote.kubernetes.secret.credentials.data["logging-tenant-id"])
	basic_auth {
		username = nonsensitive(remote.kubernetes.secret.credentials.data["logging-username"])
		password = remote.kubernetes.secret.credentials.data["logging-password"]
	}
	tls_config {
		insecure_skip_verify = ` + fmt.Sprintf("%v", lc.IsInsecureCA()) + `
	}
	}
	external_labels = {
		installation = "` + lc.GetInstallationName() + `",
		cluster_id = "` + lc.GetClusterName() + `",
		scrape_job = "kubernetes-events",
	}
}`,
				},
			},
		},
	}

	v, err := yaml.Marshal(values)
	if err != nil {
		return v1.ConfigMap{}, errors.WithStack(err)
	}

	configmap := v1.ConfigMap{
		ObjectMeta: ConfigMeta(lc),
		Data: map[string]string{
			"values": string(v),
		},
	}

	return configmap, nil
}
