package loggingconfig

import (
	"bytes"
	_ "embed"
	"fmt"
	"slices"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/blang/semver"
	capi "sigs.k8s.io/cluster-api/api/core/v1beta1" //nolint:staticcheck // SA1019 deprecated package

	"github.com/giantswarm/logging-operator/pkg/common"
)

var (
	//go:embed alloy/logging.alloy.template
	alloyLogging         string
	alloyLoggingTemplate *template.Template

	//go:embed alloy/logging-config.alloy.yaml.template
	alloyLoggingConfig         string
	alloyLoggingConfigTemplate *template.Template

	alloyNodeFilterFixedObservabilityBundleAppVersion = semver.MustParse("2.4.0")
	alloyNodeFilterImageVersion                       = semver.MustParse("1.12.0")
)

func init() {
	alloyLoggingTemplate = template.Must(template.New("logging.alloy").Funcs(sprig.FuncMap()).Parse(alloyLogging))
	alloyLoggingConfigTemplate = template.Must(template.New("logging-config.alloy.yaml").Funcs(sprig.FuncMap()).Parse(alloyLoggingConfig))
}

// GenerateAlloyLoggingConfig returns a configmap for
// the logging extra-config
func GenerateAlloyLoggingConfig(cluster *capi.Cluster, observabilityBundleVersion semver.Version, defaultNamespaces, tenants []string, clusterLabels common.ClusterLabels, insecureCA bool, enableNodeFiltering bool) (string, error) {
	var values bytes.Buffer

	alloyConfig, err := generateAlloyConfig(tenants, clusterLabels, insecureCA, enableNodeFiltering)
	if err != nil {
		return "", err
	}

	data := struct {
		AlloyConfig                      string
		AlloyImageTag                    *string
		DefaultWorkloadClusterNamespaces []string
		DefaultWriteTenant               string
		NodeFilteringEnabled             bool
		IsWorkloadCluster                bool
		PriorityClassName                string
	}{
		AlloyConfig:                      alloyConfig,
		DefaultWorkloadClusterNamespaces: defaultNamespaces,
		DefaultWriteTenant:               common.DefaultWriteTenant,
		NodeFilteringEnabled:             enableNodeFiltering,
		IsWorkloadCluster:                common.IsWorkloadCluster(clusterLabels.Installation, clusterLabels.ClusterID),
		PriorityClassName:                common.PriorityClassName,
	}

	if enableNodeFiltering && observabilityBundleVersion.LT(alloyNodeFilterFixedObservabilityBundleAppVersion) {
		// Use fixed image version
		imageTag := fmt.Sprintf("v%s", alloyNodeFilterImageVersion.String())
		data.AlloyImageTag = &imageTag
	}

	if enableNodeFiltering && observabilityBundleVersion.LT(alloyNodeFilterFixedObservabilityBundleAppVersion) {
		// Use fixed image version
		imageTag := fmt.Sprintf("v%s", alloyNodeFilterImageVersion.String())
		data.AlloyImageTag = &imageTag
	}

	err = alloyLoggingConfigTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}

func generateAlloyConfig(tenants []string, clusterLabels common.ClusterLabels, insecureCA bool, enableNodeFiltering bool) (string, error) {
	var values bytes.Buffer

	// Ensure default tenant is included in the list of tenants
	if !slices.Contains(tenants, common.DefaultWriteTenant) {
		tenants = append(tenants, common.DefaultWriteTenant)
	}

	data := struct {
		ClusterID            string
		ClusterType          string
		Organization         string
		Provider             string
		MaxBackoffPeriod     string
		RemoteTimeout        string
		IsWorkloadCluster    bool
		NodeFilteringEnabled bool
		InsecureSkipVerify   bool
		SecretName           string
		LoggingURLKey        string
		LoggingTenantIDKey   string
		LoggingUsernameKey   string
		LoggingPasswordKey   string
		LokiRulerAPIURLKey   string
		Tenants              []string
	}{
		ClusterID:            clusterLabels.ClusterID,
		ClusterType:          clusterLabels.ClusterType,
		Organization:         clusterLabels.Organization,
		Provider:             clusterLabels.Provider,
		MaxBackoffPeriod:     common.LokiMaxBackoffPeriod.String(),
		RemoteTimeout:        common.LokiRemoteTimeout.String(),
		IsWorkloadCluster:    common.IsWorkloadCluster(clusterLabels.Installation, clusterLabels.ClusterID),
		NodeFilteringEnabled: enableNodeFiltering,
		InsecureSkipVerify:   insecureCA,
		SecretName:           common.AlloyLogAgentAppName,
		LoggingURLKey:        common.LoggingURL,
		LoggingTenantIDKey:   common.LoggingTenantID,
		LoggingUsernameKey:   common.LoggingUsername,
		LoggingPasswordKey:   common.LoggingPassword,
		LokiRulerAPIURLKey:   common.LokiRulerAPIURL,
		Tenants:              tenants,
	}

	if err := alloyLoggingTemplate.Execute(&values, data); err != nil {
		return "", err
	}

	return values.String(), nil
}
