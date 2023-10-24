package grafanaagentconfig

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
	grafanaAgentConfigName = "grafana-agent-config"
)

// /// Grafana-Agent values config structure
type values struct {
	GrafanaAgent grafanaAgent `yaml:"grafana-agent" json:"grafana-agent"`
}

type grafanaAgent struct {
	Agent      agent      `yaml:"agent" json:"agent"`
	Controller controller `yaml:"controller" json:"controller"`
}

type agent struct {
	ConfigMap configMap `yaml:"configMap" json:"configMap"`
}

type configMap struct {
	Content string `yaml:"content" json:"content"`
}

type controller struct {
	InitContainers []initContainers `yaml:"initContainers" json:"initContainers"`
}

type initContainers struct {
	Name    string   `yaml:"name" json:"name"`
	Image   string   `yaml:"image" json:"image"`
	Command []string `yaml:"command" json:"command"`
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
func GenerateGrafanaAgentConfig(lc loggedcluster.Interface, credentialsSecret *v1.Secret, lokiURL string) (v1.ConfigMap, error) {

	clusterName := lc.GetClusterName()
	writeUser := clusterName
	writePassword, err := loggingcredentials.GetPassword(lc, credentialsSecret, clusterName)
	if err != nil {
		return v1.ConfigMap{}, errors.WithStack(err)
	}

	namespacesScraped := "[]"
	if common.IsWorkloadCluster(lc) {
		namespacesScraped = "[\"kube-system\", \"giantswarm\"]"
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
	namespaces = ` + namespacesScraped + `
	forward_to = [loki.write.default.receiver]
}

loki.write "default" {
	endpoint {
	url = "` + fmt.Sprintf("https://%s/loki/api/v1/push", lokiURL) + `"
	tenant_id = "` + clusterName + `"
	basic_auth {
		username = "` + writeUser + `"
		password_file = "/etc/agent/logging-write-secret"
	}
	}
	external_labels = {
		installation = "` + lc.GetInstallationName() + `",
		cluster_id = "` + clusterName + `",
		scrape_job = "kubernetes-events",
	}
}`,
				},
			},
			Controller: controller{
				InitContainers: []initContainers{
					{
						Name:  "store-logging-write-password",
						Image: "busybox:1.36",
						Command: []string{
							"- sh",
							"- -c",
							"- echo -n " + writePassword + " > /etc/agent/logging-write-secret",
						},
					},
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
