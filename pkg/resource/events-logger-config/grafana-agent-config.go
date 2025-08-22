package eventsloggerconfig

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/logging-operator/pkg/common"
	eventsloggersecret "github.com/giantswarm/logging-operator/pkg/resource/events-logger-secret"
)

var (
	//go:embed grafana-agent/events-logger.grafanaagent.template
	grafanaAgent         string
	grafanaAgentTemplate *template.Template

	//go:embed grafana-agent/events-logger-config.grafanaagent.yaml.template
	grafanaAgentConfig         string
	grafanaAgentConfigTemplate *template.Template
)

func init() {
	grafanaAgentTemplate = template.Must(template.New("events-logger.grafanaagent").Funcs(sprig.FuncMap()).Parse(grafanaAgent))
	grafanaAgentConfigTemplate = template.Must(template.New("events-logger-config.grafanaagent.yaml").Funcs(sprig.FuncMap()).Parse(grafanaAgentConfig))
}

// generateGrafanaAgentConfig returns a configmap for
// the grafana-agent extra-config
func generateGrafanaAgentConfig(cluster *capi.Cluster, loggingAgent *common.LoggingAgent, includeNamespaces []string, excludeNamespaces []string, installationName string, insecureCA bool) (string, error) {
	var values bytes.Buffer

	grafanaAgentInnerConfig, err := generateGrafanaAgentInnerConfig(cluster, loggingAgent, includeNamespaces, excludeNamespaces, installationName, insecureCA)
	if err != nil {
		return "", err
	}

	data := struct {
		GrafanaAgentInnerConfig string
	}{
		GrafanaAgentInnerConfig: grafanaAgentInnerConfig,
	}

	err = grafanaAgentConfigTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}

func generateGrafanaAgentInnerConfig(cluster *capi.Cluster, loggingAgent *common.LoggingAgent, includeNamespaces []string, excludeNamespaces []string, installationName string, insecureCA bool) (string, error) {
	var values bytes.Buffer

	data := struct {
		ClusterID          string
		Installation       string
		InsecureSkipVerify string
		RemoteTimeout      string
		SecretName         string
		IncludeNamespaces  []string
		ExcludeNamespaces  []string
		LoggingURLKey      string
		LoggingTenantIDKey string
		LoggingUsernameKey string
		LoggingPasswordKey string
		IsWorkloadCluster  bool
	}{
		ClusterID:          cluster.GetName(),
		Installation:       installationName,
		InsecureSkipVerify: fmt.Sprintf("%t", insecureCA),
		RemoteTimeout:      common.LokiRemoteTimeout.String(),
		SecretName:         eventsloggersecret.GetEventsLoggerSecretName(cluster, loggingAgent),
		IncludeNamespaces:  includeNamespaces,
		ExcludeNamespaces:  excludeNamespaces,
		LoggingURLKey:      common.LoggingURL,
		LoggingTenantIDKey: common.LoggingTenantID,
		LoggingUsernameKey: common.LoggingUsername,
		LoggingPasswordKey: common.LoggingPassword,
		IsWorkloadCluster:  common.IsWorkloadCluster(installationName, cluster.GetName()),
	}

	err := grafanaAgentTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}
