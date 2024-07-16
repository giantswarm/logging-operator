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
	alloyLogging string

	//go:embed alloy/logging-secret.yaml.template
	alloyLoggingSecret string
)

func GenerateAlloyLoggingSecret(lc loggedcluster.Interface, credentialsSecret *v1.Secret, lokiURL string) ([]byte, error) {
	var values bytes.Buffer

	t, err := template.New("logging-config.yaml").Funcs(sprig.FuncMap()).Parse(alloyLoggingSecret)
	if err != nil {
		return nil, err
	}

	alloyConfig, err := generateAlloyConfig(lc, credentialsSecret, lokiURL)
	if err != nil {
		return nil, err
	}

	data := struct{ AlloyConfig string }{
		AlloyConfig: alloyConfig,
	}

	err = t.Execute(&values, data)
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

	t, err := template.New("logging.alloy").Funcs(sprig.FuncMap()).Parse(alloyLogging)
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

	err = t.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}
