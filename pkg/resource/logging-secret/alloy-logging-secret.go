package loggingsecret

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"text/template"

	capi "sigs.k8s.io/cluster-api/api/core/v1beta1" //nolint:staticcheck // SA1019 deprecated package

	"github.com/Masterminds/sprig/v3"
	"github.com/giantswarm/observability-operator/pkg/auth"

	"github.com/giantswarm/logging-operator/pkg/common"
)

var (
	//go:embed alloy/alloy-secret.yaml.template
	alloySecret         string
	alloySecretTemplate *template.Template
)

func init() {
	alloySecretTemplate = template.Must(template.New("logging-secret.yaml").Funcs(sprig.FuncMap()).Parse(alloySecret))
}

func GenerateAlloyLoggingSecret(ctx context.Context, cluster *capi.Cluster, logsAuthManager, tracesAuthManager auth.AuthManager, lokiURL string, tracingEnabled bool) (map[string][]byte, error) {
	clusterName := cluster.GetName()
	var values bytes.Buffer

	writePassword, err := logsAuthManager.GetClusterPassword(ctx, cluster)
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

	if tracingEnabled {
		tracingPassword, err := tracesAuthManager.GetClusterPassword(ctx, cluster)
		if err != nil {
			return nil, err
		}

		templateData.ExtraSecretEnv[common.TracingUsername] = clusterName
		templateData.ExtraSecretEnv[common.TracingPassword] = tracingPassword
	}

	err = alloySecretTemplate.Execute(&values, templateData)
	if err != nil {
		return nil, err
	}

	data := make(map[string][]byte)
	data["values"] = values.Bytes()

	return data, nil
}
