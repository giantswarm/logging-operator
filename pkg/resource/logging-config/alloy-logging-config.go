package loggingconfig

import (
	"bytes"
	_ "embed"
	"text/template"

	"github.com/Masterminds/sprig/v3"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	loggingsecret "github.com/giantswarm/logging-operator/pkg/resource/logging-secret"
)

var (
	//go:embed alloy/logging.alloy.template
	alloyLogging         string
	alloyLoggingTemplate *template.Template

	//go:embed alloy/logging-config.alloy.yaml.template
	alloyLoggingConfig         string
	alloyLoggingConfigTemplate *template.Template
)

func init() {
	alloyLoggingTemplate = template.Must(template.New("logging.alloy").Funcs(sprig.FuncMap()).Parse(alloyLogging))
	alloyLoggingConfigTemplate = template.Must(template.New("logging-config.alloy.yaml").Funcs(sprig.FuncMap()).Parse(alloyLoggingConfig))
}

// GenerateAlloyLoggingConfig returns a configmap for
// the logging extra-config
func GenerateAlloyLoggingConfig(lc loggedcluster.Interface, defaultWorkloadClusterNamespaces []string) (string, error) {
	var values bytes.Buffer

	alloyConfig, err := generateAlloyConfig(lc)
	if err != nil {
		return "", err
	}

	data := struct {
		AlloyConfig                      string
		DefaultWorkloadClusterNamespaces []string
		IsWorkloadCluster                bool
		SecretName                       string
	}{
		AlloyConfig:                      alloyConfig,
		DefaultWorkloadClusterNamespaces: defaultWorkloadClusterNamespaces,
		IsWorkloadCluster:                common.IsWorkloadCluster(lc),
		SecretName:                       common.AlloyLogAgentAppName,
	}

	err = alloyLoggingConfigTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}

func generateAlloyConfig(lc loggedcluster.Interface) (string, error) {
	var values bytes.Buffer

	clusterName := lc.GetClusterName()

	data := struct {
		ClusterID                   string
		Installation                string
		MaxBackoffPeriod            string
		IsWorkloadCluster           bool
		LokiURLEnvVarName           string
		TenantIDEnvVarName          string
		BasicAuthUsernameEnvVarName string
		BasicAuthPasswordEnvVarName string
	}{
		ClusterID:                   clusterName,
		Installation:                lc.GetInstallationName(),
		MaxBackoffPeriod:            common.MaxBackoffPeriod,
		IsWorkloadCluster:           common.IsWorkloadCluster(lc),
		LokiURLEnvVarName:           loggingsecret.AlloyLokiURLEnvVarName,
		TenantIDEnvVarName:          loggingsecret.AlloyTenantIDEnvVarName,
		BasicAuthUsernameEnvVarName: loggingsecret.AlloyBasicAuthUsernameEnvVarName,
		BasicAuthPasswordEnvVarName: loggingsecret.AlloyBasicAuthPasswordEnvVarName,
	}

	err := alloyLoggingTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}
