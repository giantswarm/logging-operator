package loggingsecret

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"

	v1 "k8s.io/api/core/v1"

	"github.com/Masterminds/sprig/v3"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"
)

var (
	//go:embed alloy/logging.alloy.template
	alloyLogging         string
	alloyLoggingTemplate *template.Template

	//go:embed alloy/logging-secret.yaml.template
	alloyLoggingSecret         string
	alloyLoggingSecretTemplate *template.Template
)

func init() {
	alloyLoggingTemplate = template.Must(template.New("logging.alloy").Funcs(sprig.FuncMap()).Parse(alloyLogging))
	alloyLoggingSecretTemplate = template.Must(template.New("logging-config.yaml").Funcs(sprig.FuncMap()).Parse(alloyLoggingSecret))
}

func GenerateAlloyLoggingSecret(lc loggedcluster.Interface, credentialsSecret *v1.Secret, lokiURL string) ([]byte, error) {
	var values bytes.Buffer

	alloyConfig, err := generateAlloyConfig(lc, credentialsSecret, lokiURL)
	if err != nil {
		return nil, err
	}

	data := struct{ AlloyConfig string }{
		AlloyConfig: alloyConfig,
	}

	err = alloyLoggingSecretTemplate.Execute(&values, data)
	if err != nil {
		return nil, err
	}

	return values.Bytes(), nil
}

func generateAlloyConfig(lc loggedcluster.Interface, credentialsSecret *v1.Secret, lokiURL string) (string, error) {
	var values bytes.Buffer

	clusterName := lc.GetClusterName()

	writePassword, err := loggingcredentials.GetPassword(lc, credentialsSecret, clusterName)
	if err != nil {
		return "", err
	}

	data := struct {
		LokiURL           string
		ClusterID         string
		Installation      string
		TenantID          string
		BasicAuthUsername string
		BasicAuthPassword string
		MaxBackoffPeriod  string
		IsWorkloadCluster bool
	}{
		LokiURL:           fmt.Sprintf(common.LokiURLFormat, lokiURL),
		ClusterID:         clusterName,
		Installation:      lc.GetInstallationName(),
		TenantID:          clusterName,
		BasicAuthUsername: clusterName,
		BasicAuthPassword: writePassword,
		MaxBackoffPeriod:  common.MaxBackoffPeriod,
		IsWorkloadCluster: common.IsWorkloadCluster(lc),
	}

	err = alloyLoggingTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}
