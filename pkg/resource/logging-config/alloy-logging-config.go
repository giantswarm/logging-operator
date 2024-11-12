package loggingconfig

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
	//go:embed alloy/logging.alloy.template
	alloyLogging         string
	alloyLoggingTemplate *template.Template

	//go:embed alloy/logging-config.alloy.yaml.template
	alloyLoggingConfig         string
	alloyLoggingConfigTemplate *template.Template

	supportPodLogs = semver.MustParse("1.7.0")
)

func init() {
	alloyLoggingTemplate = template.Must(template.New("logging.alloy").Funcs(sprig.FuncMap()).Parse(alloyLogging))
	alloyLoggingConfigTemplate = template.Must(template.New("logging-config.alloy.yaml").Funcs(sprig.FuncMap()).Parse(alloyLoggingConfig))
}

// GenerateAlloyLoggingConfig returns a configmap for
// the logging extra-config
func GenerateAlloyLoggingConfig(lc loggedcluster.Interface, observabilityBundleVersion semver.Version, defaultNamespaces []string) (string, error) {
	var values bytes.Buffer

	alloyConfig, err := generateAlloyConfig(lc, observabilityBundleVersion)
	if err != nil {
		return "", err
	}

	data := struct {
		AlloyConfig                      string
		DefaultWorkloadClusterNamespaces []string
		IsWorkloadCluster                bool
		SecretName                       string
		SupportPodLogs                   bool
		LogsEvents                       bool
	}{
		AlloyConfig:                      alloyConfig,
		DefaultWorkloadClusterNamespaces: defaultNamespaces,
		IsWorkloadCluster:                common.IsWorkloadCluster(lc),
		SecretName:                       common.AlloyLogAgentAppName,
		// Observability bundle in older versions do not support PodLogs
		SupportPodLogs: observabilityBundleVersion.GE(supportPodLogs),
		LogsEvents:     true,
	}

	err = alloyLoggingConfigTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}

func generateAlloyConfig(lc loggedcluster.Interface, observabilityBundleVersion semver.Version) (string, error) {
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
		SupportPodLogs              bool
	}{
		ClusterID:                   clusterName,
		Installation:                lc.GetInstallationName(),
		MaxBackoffPeriod:            common.MaxBackoffPeriod,
		IsWorkloadCluster:           common.IsWorkloadCluster(lc),
		LokiURLEnvVarName:           loggingsecret.AlloyLokiURLEnvVarName,
		TenantIDEnvVarName:          loggingsecret.AlloyTenantIDEnvVarName,
		BasicAuthUsernameEnvVarName: loggingsecret.AlloyBasicAuthUsernameEnvVarName,
		BasicAuthPasswordEnvVarName: loggingsecret.AlloyBasicAuthPasswordEnvVarName,
		// Observability bundle in older versions do not support PodLogs
		SupportPodLogs: observabilityBundleVersion.GE(supportPodLogs),
	}

	err := alloyLoggingTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}
