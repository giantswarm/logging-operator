package loggingconfig

import (
	"bytes"
	_ "embed"
	"slices"
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
func GenerateAlloyLoggingConfig(lc loggedcluster.Interface, observabilityBundleVersion semver.Version, defaultNamespaces, tenants []string) (string, error) {
	var values bytes.Buffer

	alloyConfig, err := generateAlloyConfig(lc, observabilityBundleVersion, tenants)
	if err != nil {
		return "", err
	}

	data := struct {
		AlloyConfig                      string
		DefaultWorkloadClusterNamespaces []string
		DefaultWriteTenant               string
		IsWorkloadCluster                bool
		SupportPodLogs                   bool
		SupportVPA                       bool
	}{
		AlloyConfig:                      alloyConfig,
		DefaultWorkloadClusterNamespaces: defaultNamespaces,
		DefaultWriteTenant:               lc.GetTenant(),
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

func generateAlloyConfig(lc loggedcluster.Interface, observabilityBundleVersion semver.Version, tenants []string) (string, error) {
	var values bytes.Buffer

	clusterName := lc.GetClusterName()

	// Ensure default tenant is included in the list of tenants
	if !slices.Contains(tenants, lc.GetTenant()) {
		tenants = append(tenants, lc.GetTenant())
	}

	data := struct {
		ClusterID          string
		Installation       string
		MaxBackoffPeriod   string
		RemoteTimeout      string
		IsWorkloadCluster  bool
		SupportPodLogs     bool
		InsecureSkipVerify bool
		SecretName         string
		LoggingURLKey      string
		LoggingTenantIDKey string
		LoggingUsernameKey string
		LoggingPasswordKey string
		LokiRulerAPIURLKey string
		Tenants            []string
	}{
		ClusterID:         clusterName,
		Installation:      lc.GetInstallationName(),
		MaxBackoffPeriod:  common.LokiMaxBackoffPeriod.String(),
		RemoteTimeout:     common.LokiRemoteTimeout.String(),
		IsWorkloadCluster: common.IsWorkloadCluster(lc),
		// Observability bundle in older versions do not support PodLogs
		SupportPodLogs:     observabilityBundleVersion.GE(supportPodLogs),
		InsecureSkipVerify: lc.IsInsecureCA(),
		SecretName:         common.AlloyLogAgentAppName,
		LoggingURLKey:      common.LoggingURL,
		LoggingTenantIDKey: common.LoggingTenantID,
		LoggingUsernameKey: common.LoggingUsername,
		LoggingPasswordKey: common.LoggingPassword,
		LokiRulerAPIURLKey: common.LokiRulerAPIURL,
		Tenants:            tenants,
	}

	err := alloyLoggingTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}
