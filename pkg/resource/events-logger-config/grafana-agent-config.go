package eventsloggerconfig

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	eventsloggersecret "github.com/giantswarm/logging-operator/pkg/resource/events-logger-secret"
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
	grafanaAgentConfigTemplate = template.Must(template.New("events-logger.grafanaagent.yaml").Funcs(sprig.FuncMap()).Parse(grafanaAgentConfig))
}

// GenerateGrafanaAgentConfig returns a configmap for
// the grafana-agent extra-config
func GenerateGrafanaAgentConfig(lc loggedcluster.Interface, defaultNamespaces []string) (string, error) {
	var values bytes.Buffer

	grafanaAgentInnerConfig, err := generateGrafanaAgentInnerConfig(lc, defaultNamespaces)
	if err != nil {
		return "", err
	}

	data := struct {
		GrafanaAgentInnerConfig string
		Replicas                int
		Type                    string
		Create                  string
	}{
		GrafanaAgentInnerConfig: grafanaAgentInnerConfig,
		Replicas:                1,
		Type:                    "deployment",
		Create:                  "false",
	}

	err = grafanaAgentConfigTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}

func generateGrafanaAgentInnerConfig(lc loggedcluster.Interface, defaultNamespaces []string) (string, error) {
	var values bytes.Buffer

	data := struct {
		ClusterID          string
		Installation       string
		InsecureSkipVerify string
		SecretName         string
		SecretNamespace    string
		ScrapedNamespaces  string
	}{
		ClusterID:          lc.GetClusterName(),
		Installation:       lc.GetInstallationName(),
		InsecureSkipVerify: fmt.Sprintf("%t", lc.IsInsecureCA()),
		SecretName:         eventsloggersecret.GetEventsLoggerSecretName(lc),
		SecretNamespace:    lc.GetAppsNamespace(),
		ScrapedNamespaces:  common.FormatScrapedNamespaces(lc, defaultNamespaces),
	}

	err := grafanaAgentTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}
