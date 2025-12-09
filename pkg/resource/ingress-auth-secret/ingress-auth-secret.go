package ingressauthsecret

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/core/v1beta1"

	"github.com/giantswarm/logging-operator/pkg/common"
	credentials "github.com/giantswarm/logging-operator/pkg/resource/credentials"
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

// listUsers returns a map of users found in a credentialsSecret
func listUsers(credentialsSecret *v1.Secret) []string {
	var usersList []string
	for myUser := range credentialsSecret.Data {
		usersList = append(usersList, myUser)
	}

	return usersList
}

// generateIngressAuthSecret returns a secret for the loki ingress auth
func generateIngressAuthSecret(cluster *capi.Cluster, credentialsSecret *v1.Secret) (map[string]string, error) {
	users := make(map[string]string)
	// Loop on write users
	for _, user := range listUsers(credentialsSecret) {
		writePassword, err := credentials.GetPassword(cluster, credentialsSecret, user)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		password, err := bcrypt.GenerateFromPassword([]byte(writePassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		users[user] = string(password)
	}

	return users, nil
}
