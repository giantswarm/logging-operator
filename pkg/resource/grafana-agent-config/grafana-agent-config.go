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

// grafana-agent:
//   image:
//     # -- Grafana Agent image repository.
//     repository: giantswarm/grafana-agent
//   configReloader:
//     image:
//       repository: giantswarm/configmap-reload

//   agent:
//     # -- Mode to run Grafana Agent in. Can be "flow" or "static".
//     mode: 'flow'
//     configMap:
//       # -- Create a new ConfigMap for the config file.
//       create: true
//       # -- Content to assign to the new ConfigMap.  This is passed into `tpl` allowing for templating from values.
//       content: |
//         logging {
//           level  = "info"
//           format = "logfmt"
//         }

///// Grafana-Agent values config structure

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
func GenerateGrafanaAgentConfig(lc loggedcluster.Interface, promtailCredentials map[string]string) (v1.ConfigMap, error) {

	values := values{
		GrafanaAgent: grafanaAgent{
			Agent: agent{
				ConfigMap: configMap{
					Content: `# grafana-agent river config
logging {
	level  = "info"
	format = "logfmt"
}

loki.source.kubernetes_events "local" {
	namespaces = ["kube-system", "giantswarm"]
	forward_to = [loki.write.default.receiver]
}

loki.write "default" {
	endpoint {
	url = "` + promtailCredentials["url"] + `"
	tenant_id = "` + promtailCredentials["tenantID"] + `"
	basic_auth {
		username = "` + promtailCredentials["username"] + `"
		password = secret("` + promtailCredentials["password"] + `")
	}
	}
	external_labels = {
		installation = "` + promtailCredentials["installation"] + `",
		cluster_id = "` + promtailCredentials["clusterID"] + `",
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
