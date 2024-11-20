package eventsloggerconfig

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	loggingsecret "github.com/giantswarm/logging-operator/pkg/resource/logging-secret"
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
	alloyEventsConfigTemplate = template.Must(template.New("events-logger.alloy.yaml").Funcs(sprig.FuncMap()).Parse(alloyEventsConfig))
}

func generateAlloyEventsConfig(lc loggedcluster.Interface, defaultNamespaces []string) (string, error) {
	var values bytes.Buffer

	alloyConfig, err := generateAlloyConfig(lc, defaultNamespaces)
	if err != nil {
		return "", err
	}

	data := struct {
		AlloyConfig string
		SecretName  string
	}{
		AlloyConfig: alloyConfig,
		SecretName:  common.AlloyEventsLoggerAppName,
	}

	err = alloyEventsConfigTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}

func generateAlloyConfig(lc loggedcluster.Interface, defaultNamespaces []string) (string, error) {
	var values bytes.Buffer

	data := struct {
		ClusterID                   string
		Installation                string
		InsecureSkipVerify          string
		MaxBackoffPeriod            string
		LokiURLEnvVarName           string
		TenantIDEnvVarName          string
		BasicAuthUsernameEnvVarName string
		BasicAuthPasswordEnvVarName string
		ScrapedNamespaces           string
	}{
		ClusterID:                   lc.GetClusterName(),
		Installation:                lc.GetInstallationName(),
		InsecureSkipVerify:          fmt.Sprintf("%t", lc.IsInsecureCA()),
		MaxBackoffPeriod:            common.MaxBackoffPeriod,
		LokiURLEnvVarName:           loggingsecret.AlloyLokiURLEnvVarName,
		TenantIDEnvVarName:          loggingsecret.AlloyTenantIDEnvVarName,
		BasicAuthUsernameEnvVarName: loggingsecret.AlloyBasicAuthUsernameEnvVarName,
		BasicAuthPasswordEnvVarName: loggingsecret.AlloyBasicAuthPasswordEnvVarName,
		ScrapedNamespaces:           common.FormatScrapedNamespaces(lc, defaultNamespaces),
	}

	err := alloyEventsTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}