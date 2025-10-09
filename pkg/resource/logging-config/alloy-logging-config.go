package loggingconfig

import (
	"bytes"
	_ "embed"
	"slices"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/blang/semver"

	"github.com/giantswarm/logging-operator/pkg/common"
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
func GenerateAlloyLoggingConfig(loggingAgent *common.LoggingAgent, observabilityBundleVersion semver.Version, defaultNamespaces, tenants []string, clusterLabels common.ClusterLabels, insecureCA bool) (string, error) {
	var values bytes.Buffer

	alloyConfig, err := generateAlloyConfig(observabilityBundleVersion, tenants, clusterLabels, insecureCA)
	if err != nil {
		return "", err
	}

	data := struct {
		AlloyConfig                      string
		DefaultWorkloadClusterNamespaces []string
		DefaultWriteTenant               string
		IsWorkloadCluster                bool
		PriorityClassName                string
		SupportPodLogs                   bool
		SupportVPA                       bool
	}{
		AlloyConfig:                      alloyConfig,
		DefaultWorkloadClusterNamespaces: defaultNamespaces,
		DefaultWriteTenant:               common.DefaultWriteTenant,
		IsWorkloadCluster:                common.IsWorkloadCluster(clusterLabels.Installation, clusterLabels.ClusterID),
		PriorityClassName:                common.PriorityClassName,
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

func generateAlloyConfig(observabilityBundleVersion semver.Version, tenants []string, clusterLabels common.ClusterLabels, insecureCA bool) (string, error) {
	var values bytes.Buffer

	// Ensure default tenant is included in the list of tenants
	if !slices.Contains(tenants, common.DefaultWriteTenant) {
		tenants = append(tenants, common.DefaultWriteTenant)
	}

	data := struct {
		ClusterID          string
		ClusterType        string
		Customer           string
		Installation       string
		Organization       string
		Pipeline           string
		Provider           string
		Region             string
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
		ClusterID:         clusterLabels.ClusterID,
		ClusterType:       clusterLabels.ClusterType,
		Customer:          clusterLabels.Customer,
		Installation:      clusterLabels.Installation,
		Organization:      clusterLabels.Organization,
		Pipeline:          clusterLabels.Pipeline,
		Provider:          clusterLabels.Provider,
		Region:            clusterLabels.Region,
		MaxBackoffPeriod:  common.LokiMaxBackoffPeriod.String(),
		RemoteTimeout:     common.LokiRemoteTimeout.String(),
		IsWorkloadCluster: common.IsWorkloadCluster(clusterLabels.Installation, clusterLabels.ClusterID),
		// Observability bundle in older versions do not support PodLogs
		SupportPodLogs:     observabilityBundleVersion.GE(supportPodLogs),
		InsecureSkipVerify: insecureCA,
		SecretName:         common.AlloyLogAgentAppName,
		LoggingURLKey:      common.LoggingURL,
		LoggingTenantIDKey: common.LoggingTenantID,
		LoggingUsernameKey: common.LoggingUsername,
		LoggingPasswordKey: common.LoggingPassword,
		LokiRulerAPIURLKey: common.LokiRulerAPIURL,
		Tenants:            tenants,
	}

	if err := alloyLoggingTemplate.Execute(&values, data); err != nil {
		return "", err
	}

	return values.String(), nil
}
