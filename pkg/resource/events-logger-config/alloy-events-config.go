package eventsloggerconfig

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
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

func generateAlloyEventsConfig(lc loggedcluster.Interface, includeNamespaces []string, excludeNamespaces []string, installationName string, insecureCA bool) (string, error) {
	var values bytes.Buffer

	alloyConfig, err := generateAlloyConfig(lc, includeNamespaces, excludeNamespaces, installationName, insecureCA)
	if err != nil {
		return "", err
	}

	data := struct {
		AlloyConfig string
	}{
		AlloyConfig: alloyConfig,
	}

	err = alloyEventsConfigTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}

func generateAlloyConfig(lc loggedcluster.Interface, includeNamespaces []string, excludeNamespaces []string, installationName string, insecureCA bool) (string, error) {
	var values bytes.Buffer

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
	}{
		ClusterID:          lc.GetClusterName(),
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
		IsWorkloadCluster:  common.IsWorkloadCluster(installationName, lc.GetClusterName()),
	}

	err := alloyEventsTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}
