package k8seventsconfig

import (
	"bytes"
	_ "embed"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/blang/semver"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	loggingsecret "github.com/giantswarm/logging-operator/pkg/resource/logging-secret"
)

var (
	//go:embed alloy/events-logger.alloy.template
	alloyEvents         string
	alloyEventsTemplate *template.Template

	//go:embed alloy/events-logger-config.alloy.yaml.template
	alloyEventsConfig         string
	alloyEventsConfigTemplate *template.Template
)

func init() {
	alloyEventsTemplate = template.Must(template.New("events-logger.alloy").Funcs(sprig.FuncMap()).Parse(alloyEvents))
	alloyEventsConfigTemplate = template.Must(template.New("events-logger.alloy.yaml").Funcs(sprig.FuncMap()).Parse(alloyEventsConfig))
}

func GenerateAlloyEventsConfig(lc loggedcluster.Interface, observabilityBundleVersion semver.Version, defaultNamespaces []string) (string, error) {
	var values bytes.Buffer

	alloyConfig, err := generateAlloyConfig(lc, defaultNamespaces)
	if err != nil {
		return "", err
	}

	data := struct {
		GrafanaAgentInnerConfig string
		Replicas                int
		Type                    string
		SecretName              string
	}{
		GrafanaAgentInnerConfig: alloyConfig,
		Replicas:                1,
		Type:                    "deployment",
		// We're using the same secret used by alloy-logs to authenticate against Loki
		SecretName: common.AlloyLogAgentAppName,
	}

	err = grafanaAgentConfigTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}

func generateAlloyConfig(lc loggedcluster.Interface, defaultNamespaces []string) (string, error) {
	var values bytes.Buffer

	data := struct {
		ClusterID                        string
		Installation                     string
		DefaultWorkloadClusterNamespaces []string
		MaxBackoffPeriod                 string
		LokiURLEnvVarName                string
		TenantIDEnvVarName               string
		BasicAuthUsernameEnvVarName      string
		BasicAuthPasswordEnvVarName      string
	}{
		ClusterID:                        lc.GetClusterName(),
		Installation:                     lc.GetInstallationName(),
		DefaultWorkloadClusterNamespaces: defaultNamespaces,
		MaxBackoffPeriod:                 common.MaxBackoffPeriod,
		LokiURLEnvVarName:                loggingsecret.AlloyLokiURLEnvVarName,
		TenantIDEnvVarName:               loggingsecret.AlloyTenantIDEnvVarName,
		BasicAuthUsernameEnvVarName:      loggingsecret.AlloyBasicAuthUsernameEnvVarName,
		BasicAuthPasswordEnvVarName:      loggingsecret.AlloyBasicAuthPasswordEnvVarName,
	}

	err := alloyEventsTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}
