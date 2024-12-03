package loggingconfig

import (
	"bytes"
	_ "embed"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/blang/semver"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

var (
	//go:embed alloy/logging.alloy.template
	alloyLogging         string
	alloyLoggingTemplate *template.Template

	//go:embed alloy/logging-config.alloy.yaml.template
	alloyLoggingConfig         string
	alloyLoggingConfigTemplate *template.Template

	supportPodLogs = semver.MustParse("1.7.0")
	supportVPA     = supportPodLogs
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
		SupportPodLogs                   bool
		SupportVPA                       bool
	}{
		AlloyConfig:                      alloyConfig,
		DefaultWorkloadClusterNamespaces: defaultNamespaces,
		IsWorkloadCluster:                common.IsWorkloadCluster(lc),
		// Observability bundle in older versions do not support PodLogs
		SupportPodLogs: observabilityBundleVersion.GE(supportPodLogs),
		// Observability bundle in older versions do not support VPA
		SupportVPA: observabilityBundleVersion.GE(supportVPA),
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
		ClusterID          string
		Installation       string
		MaxBackoffPeriod   string
		IsWorkloadCluster  bool
		SupportPodLogs     bool
		InsecureSkipVerify bool
		SecretName         string
		LoggingURLKey      string
		LoggingTenantIDKey string
		LoggingUsernameKey string
		LoggingPasswordKey string
	}{
		ClusterID:         clusterName,
		Installation:      lc.GetInstallationName(),
		MaxBackoffPeriod:  common.MaxBackoffPeriod,
		IsWorkloadCluster: common.IsWorkloadCluster(lc),
		// Observability bundle in older versions do not support PodLogs
		SupportPodLogs:     observabilityBundleVersion.GE(supportPodLogs),
		InsecureSkipVerify: lc.IsInsecureCA(),
		SecretName:         common.AlloyLogAgentAppName,
		LoggingURLKey:      common.LoggingURL,
		LoggingTenantIDKey: common.LoggingTenantID,
		LoggingUsernameKey: clusterName,
		LoggingPasswordKey: common.LoggingPassword,
	}

	err := alloyLoggingTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}
