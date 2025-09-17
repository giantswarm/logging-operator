package tracingsecret

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/Masterminds/sprig/v3"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"
)

const (
	TracingSecretName = "events-logger-secret" // #nosec G101
)

var (
	//go:embed tracing-secret.yaml.template
	alloySecret         string
	alloySecretTemplate *template.Template
)

func init() {
	alloySecretTemplate = template.Must(template.New("logging-secret.yaml").Funcs(sprig.FuncMap()).Parse(alloySecret))
}

func generateTracingSecret(cluster *capi.Cluster) (v1.Secret, error) {
	clusterName := cluster.GetName()

	password, err := loggingcredentials.GeneratePassword()
	if err != nil {
		return v1.Secret{}, err
	}

	templateData := struct {
		ExtraSecretEnv map[string]string
	}{
		ExtraSecretEnv: map[string]string{
			common.TracingUsername: clusterName,
			common.TracingPassword: password,
		},
	}

	var values bytes.Buffer
	err = alloySecretTemplate.Execute(&values, templateData)
	if err != nil {
		return v1.Secret{}, err
	}

	data := make(map[string][]byte)
	data["values"] = values.Bytes()

	secret := v1.Secret{
		ObjectMeta: secretMeta(cluster),
		Data:       data,
	}

	return secret, nil
}

// SecretMeta returns metadata for the events-logger-secret
func secretMeta(cluster *capi.Cluster) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      GetTracingSecretName(cluster),
		Namespace: cluster.GetNamespace(),
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

func GetTracingSecretName(cluster *capi.Cluster) string {
	return fmt.Sprintf("%s-%s", cluster.GetName(), TracingSecretName)
}
