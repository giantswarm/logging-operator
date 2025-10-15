package eventsloggerconfig

import (
	"bytes"
	_ "embed"
	"fmt"
	"net"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

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

func generateAlloyEventsConfig(cluster *capi.Cluster, includeNamespaces []string, excludeNamespaces []string, installationName string, insecureCA bool, tracingEnabled bool, tempoURL string, tenants []string) (string, error) {
	var values bytes.Buffer

	alloyConfig, err := generateAlloyConfig(cluster, includeNamespaces, excludeNamespaces, installationName, insecureCA, tracingEnabled, tempoURL, tenants)
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
		IsWorkloadCluster: common.IsWorkloadCluster(installationName, cluster.GetName()),
	}

	err = alloyEventsConfigTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}

func generateAlloyConfig(cluster *capi.Cluster, includeNamespaces []string, excludeNamespaces []string, installationName string, insecureCA bool, tracingEnabled bool, tempoURL string, tenants []string) (string, error) {
	var values bytes.Buffer

	// endpoint must be in host:port format which is required by the gRPC exporter.
	// Direclty adding port 443 here as tempoURL is an ingress host
	// which does not contain any port and exposes port 443.
	// see https://kubernetes.io/docs/reference/generated/kubernetes-api
	endpoint := net.JoinHostPort(tempoURL, "443")

	data := struct {
		ClusterID          string
		Installation       string
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
		ClusterID:          cluster.GetName(),
		Installation:       installationName,
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
		IsWorkloadCluster:  common.IsWorkloadCluster(installationName, cluster.GetName()),
		TracingEnabled:     tracingEnabled,
		TracingEndpoint:    endpoint,
		TracingUsernameKey: common.TracingUsername,
		TracingPasswordKey: common.TracingPassword,
		Tenants:            tenants,
	}

	err := alloyEventsTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}
