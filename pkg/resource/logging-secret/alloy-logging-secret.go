package loggingsecret

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"

	v1 "k8s.io/api/core/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/Masterminds/sprig/v3"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"
)

var (
	//go:embed alloy/logging-secret.yaml.template
	alloySecret         string
	alloySecretTemplate *template.Template
)

func init() {
	alloySecretTemplate = template.Must(template.New("logging-secret.yaml").Funcs(sprig.FuncMap()).Parse(alloySecret))
}

func GenerateAlloyLoggingSecret(cluster *capi.Cluster, credentialsSecret *v1.Secret, lokiURL string) (map[string][]byte, error) {
	clusterName := cluster.GetName()

	writePassword, err := loggingcredentials.GetPassword(cluster, credentialsSecret, clusterName)
	if err != nil {
		return nil, err
	}

	templateData := struct {
		ExtraSecretEnv map[string]string
	}{
		ExtraSecretEnv: map[string]string{
			common.LoggingURL:      fmt.Sprintf(common.LokiPushURLFormat, lokiURL),
			common.LoggingTenantID: common.DefaultWriteTenant,
			common.LoggingUsername: clusterName,
			common.LoggingPassword: writePassword,
			common.LokiRulerAPIURL: fmt.Sprintf(common.LokiBaseURLFormat, lokiURL),
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
