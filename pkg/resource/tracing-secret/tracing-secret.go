package tracingsecret

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"
)

const (
	TracingSecretName = "tracing-secret" // #nosec G101
)

func generateTracingSecret(cluster *capi.Cluster) (v1.Secret, error) {
	clusterName := cluster.GetName()

	password, err := loggingcredentials.GeneratePassword()
	if err != nil {
		return v1.Secret{}, err
	}

	templateData := common.AlloySecretTemplateData{
		ExtraSecretEnv: map[string]string{
			common.TracingUsername: clusterName,
			common.TracingPassword: password,
		},
	}

	values, err := common.GenerateAlloySecretValues(templateData)
	if err != nil {
		return v1.Secret{}, err
	}

	data := make(map[string][]byte)
	data["values"] = values

	secret := v1.Secret{
		ObjectMeta: secretMeta(cluster),
		Data:       data,
	}

	return secret, nil
}

// SecretMeta returns metadata for the tracing secret
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
