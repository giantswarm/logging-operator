package ingressauthsecret

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/logging-operator/pkg/common"
)

// ingressAuthSecretMetadata returns metadata for the ingresses auth secret metadata
func ingressAuthSecretMetadata(secretName string, secretNamespace string) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      secretName,
		Namespace: secretNamespace,
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

func ingressAuthSecret(secretName string, secretNamespace string) v1.Secret {
	return v1.Secret{
		ObjectMeta: ingressAuthSecretMetadata(secretName, secretNamespace),
	}
}
