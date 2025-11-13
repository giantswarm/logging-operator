package eventsloggerconfig

import (
	"bytes"
	_ "embed"
	"fmt"
	"net"
	"text/template"

	"github.com/Masterminds/sprig/v3"

	"github.com/giantswarm/logging-operator/pkg/common"
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
	alloyEventsConfigTemplate = template.Must(template.New("events-logger-config.alloy.yaml").Funcs(sprig.FuncMap()).Parse(alloyEventsConfig))
}

func generateAlloyEventsConfig(includeNamespaces, excludeNamespaces []string, insecureCA, tracingEnabled bool, tempoURL string, tenants []string, clusterLabels common.ClusterLabels) (string, error) {
	var values bytes.Buffer

	alloyConfig, err := generateAlloyConfig(includeNamespaces, excludeNamespaces, insecureCA, tracingEnabled, tempoURL, tenants, clusterLabels)
	if err != nil {
		return "", err
	}

	data := struct {
		AlloyConfig       string
		TracingEnabled    bool
		IsWorkloadCluster bool
	}{
		AlloyConfig:       alloyConfig,
		TracingEnabled:    tracingEnabled,
		IsWorkloadCluster: common.IsWorkloadCluster(clusterLabels.Installation, clusterLabels.ClusterID),
	}

	err = alloyEventsConfigTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}

func generateAlloyConfig(includeNamespaces, excludeNamespaces []string, insecureCA, tracingEnabled bool, tempoURL string, tenants []string, clusterLabels common.ClusterLabels) (string, error) {
	var values bytes.Buffer

	// endpoint must be in host:port format which is required by the gRPC exporter.
	// Direclty adding port 443 here as tempoURL is an ingress host
	// which does not contain any port and exposes port 443.
	// see https://kubernetes.io/docs/reference/generated/kubernetes-api
	endpoint := net.JoinHostPort(tempoURL, "443")

	data := struct {
		ClusterID          string
		ClusterType        string
		Organization       string
		Provider           string
		InsecureSkipVerify string
		MaxBackoffPeriod   string
		RemoteTimeout      string
		IncludeNamespaces  []string
		ExcludeNamespaces  []string
		SecretName         string
		LoggingURLKey      string
		LoggingTenantIDKey string
		LoggingUsernameKey string
		LoggingPasswordKey string
		IsWorkloadCluster  bool
		TracingEnabled     bool
		TracingEndpoint    string
		TracingUsernameKey string
		TracingPasswordKey string
		Tenants            []string
	}{
		ClusterID:          clusterLabels.ClusterID,
		ClusterType:        clusterLabels.ClusterType,
		Organization:       clusterLabels.Organization,
		Provider:           clusterLabels.Provider,
		InsecureSkipVerify: fmt.Sprintf("%t", insecureCA),
		MaxBackoffPeriod:   common.LokiMaxBackoffPeriod.String(),
		RemoteTimeout:      common.LokiRemoteTimeout.String(),
		SecretName:         common.AlloyEventsLoggerAppName,
		IncludeNamespaces:  includeNamespaces,
		ExcludeNamespaces:  excludeNamespaces,
		LoggingURLKey:      common.LoggingURL,
		LoggingTenantIDKey: common.LoggingTenantID,
		LoggingUsernameKey: common.LoggingUsername,
		LoggingPasswordKey: common.LoggingPassword,
		IsWorkloadCluster:  common.IsWorkloadCluster(clusterLabels.Installation, clusterLabels.ClusterID),
		TracingEnabled:     tracingEnabled,
		TracingEndpoint:    endpoint,
		TracingUsernameKey: common.TracingUsername,
		TracingPasswordKey: common.TracingPassword,
		Tenants:            tenants,
	}

	if err := alloyEventsTemplate.Execute(&values, data); err != nil {
		return "", err
	}

	return values.String(), nil
}
