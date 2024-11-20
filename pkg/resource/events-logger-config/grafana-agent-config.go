package eventsloggerconfig

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

const (
	grafanaAgentConfigName = "grafana-agent-config"
)

var (
	//go:embed grafana-agent/events-logger.grafanaagent.template
	grafanaAgent         string
	grafanaAgentTemplate *template.Template

	//go:embed grafana-agent/events-logger-config.grafanaagent.yaml.template
	grafanaAgentConfig         string
	grafanaAgentConfigTemplate *template.Template
)

func init() {
	grafanaAgentTemplate = template.Must(template.New("events-logger.grafanaagent").Funcs(sprig.FuncMap()).Parse(grafanaAgent))
	grafanaAgentConfigTemplate = template.Must(template.New("events-logger-config.grafanaagent.yaml").Funcs(sprig.FuncMap()).Parse(grafanaAgentConfig))
}

// configMeta returns metadata for the grafana-agent-config
func configMeta(lc loggedcluster.Interface) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      getGrafanaAgentConfigName(lc),
		Namespace: lc.GetAppsNamespace(),
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

// generateGrafanaAgentConfig returns a configmap for
// the grafana-agent extra-config
func generateGrafanaAgentConfig(lc loggedcluster.Interface, defaultWorkloadClusterNamespaces []string) (string, error) {
	var values bytes.Buffer

	grafanaAgentInnerConfig, err := generateGrafanaAgentInnerConfig(lc, defaultWorkloadClusterNamespaces)
	if err != nil {
		return "", err
	}

	data := struct {
		GrafanaAgentInnerConfig string
	}{
		GrafanaAgentInnerConfig: grafanaAgentInnerConfig,
	}

	err = grafanaAgentTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}

func generateGrafanaAgentInnerConfig(lc loggedcluster.Interface, defaultWorkloadClusterNamespaces []string) (string, error) {
	var values bytes.Buffer

	data := struct {
		ClusterID          string
		Installation       string
		InsecureSkipVerify string
		SecretName         string
		ScrapedNamespaces  string
	}{
		ClusterID:          lc.GetClusterName(),
		Installation:       lc.GetInstallationName(),
		InsecureSkipVerify: fmt.Sprintf("%t", lc.IsInsecureCA()),
		SecretName:         fmt.Sprintf("%s-%s", lc.GetClusterName(), common.GrafanaAgentExtraSecretName()),
		ScrapedNamespaces:  common.FormatScrapedNamespaces(lc, defaultWorkloadClusterNamespaces),
	}

	err := grafanaAgentConfigTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}

func getGrafanaAgentConfigName(lc loggedcluster.Interface) string {
	return fmt.Sprintf("%s-%s", lc.GetClusterName(), grafanaAgentConfigName)
}
