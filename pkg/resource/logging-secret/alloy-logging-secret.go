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

const (
	AlloyLokiURLEnvVarName           = "LOKI_URL"
	AlloyTenantIDEnvVarName          = "TENANT_ID"
	AlloyBasicAuthUsernameEnvVarName = "BASIC_AUTH_USERNAME"
	AlloyBasicAuthPasswordEnvVarName = "BASIC_AUTH_PASSWORD" // #nosec G101
)

var (
	//go:embed alloy/logging-secret.yaml.template
	alloySecret         string
	alloySecretTemplate *template.Template
)

func init() {
	alloySecretTemplate = template.Must(template.New("logging-secret.yaml").Funcs(sprig.FuncMap()).Parse(alloySecret))
}

func GenerateAlloyLoggingSecret(lc loggedcluster.Interface, credentialsSecret *v1.Secret, lokiURL string) (map[string][]byte, error) {
	clusterName := lc.GetClusterName()

	writePassword, err := loggingcredentials.GetPassword(lc, credentialsSecret, clusterName)
	if err != nil {
		return nil, err
	}

	templateData := struct {
		ExtraSecretEnv map[string]string
	}{
		ExtraSecretEnv: map[string]string{
			AlloyLokiURLEnvVarName:           fmt.Sprintf(common.LokiURLFormat, lokiURL),
			AlloyTenantIDEnvVarName:          clusterName,
			AlloyBasicAuthUsernameEnvVarName: clusterName,
			AlloyBasicAuthPasswordEnvVarName: writePassword,
		},
	}

	var values bytes.Buffer
	err = alloySecretTemplate.Execute(&values, templateData)
	if err != nil {
		return nil, err
	}

	data := make(map[string][]byte)
	data["values"] = values.Bytes()

	return data, nil
}
