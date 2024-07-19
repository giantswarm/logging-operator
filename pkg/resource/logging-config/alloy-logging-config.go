package loggingconfig

import (
	"bytes"
	_ "embed"
	"html/template"

	"github.com/Masterminds/sprig/v3"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	alloysecret "github.com/giantswarm/logging-operator/pkg/resource/alloy-secret"
)

var (
	//go:embed alloy/logging.alloy.template
	alloyLogging         string
	alloyLoggingTemplate *template.Template

	//go:embed alloy/logging-config.yaml.template
	alloyLoggingConfig         string
	alloyLoggingConfigTemplate *template.Template
)

func init() {
	alloyLoggingTemplate = template.Must(template.New("logging.alloy").Funcs(sprig.FuncMap()).Parse(alloyLogging))
	alloyLoggingConfigTemplate = template.Must(template.New("logging-config.yaml").Funcs(sprig.FuncMap()).Parse(alloyLoggingConfig))
}

// GenerateAlloyLoggingConfig returns a configmap for
// the logging extra-config
func GenerateAlloyLoggingConfig(lc loggedcluster.Interface) (string, error) {
	var values bytes.Buffer

	alloyConfig, err := generateAlloyConfig(lc)
	if err != nil {
		return "", err
	}

	data := struct {
		AlloyConfig string
		SecretName  string
	}{
		AlloyConfig: alloyConfig,
		SecretName:  alloysecret.SecretName,
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
		LokiURLEnvVarName:           alloysecret.AlloyLokiURLEnvVarName,
		TenantIDEnvVarName:          alloysecret.AlloyTenantIDEnvVarName,
		BasicAuthUsernameEnvVarName: alloysecret.AlloyBasicAuthUsernameEnvVarName,
		BasicAuthPasswordEnvVarName: alloysecret.AlloyBasicAuthPasswordEnvVarName,
	}

	err := alloyLoggingTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}
