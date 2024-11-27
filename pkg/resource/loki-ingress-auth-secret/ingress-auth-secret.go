package lokiingressauthsecret

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"
)

const (
	//#nosec G101
	lokiIngressAuthSecretName      = "loki-ingress-auth"
	lokiIngressAuthSecretNamespace = "loki"
	// DefaultReadOrgIDs - make sure to have at least 2 tenants, to prevent writing with this user
	DefaultReadOrgIDs = "giantswarm|default"
)

// lokiIngressAuthSecretMetadata returns metadata for the loki ingress auth secret metadata
func lokiIngressAuthSecretMetadata() metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      lokiIngressAuthSecretName,
		Namespace: lokiIngressAuthSecretNamespace,
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

func lokiIngressAuthSecret() v1.Secret {
	return v1.Secret{
		ObjectMeta: lokiIngressAuthSecretMetadata(),
	}
}

// listWriteUsers returns a map of users found in a credentialsSecret
func listWriteUsers(credentialsSecret *v1.Secret) []string {
	var usersList []string
	for myUser := range credentialsSecret.Data {
		// bypass read user
		if myUser != common.ReadUser {
			usersList = append(usersList, myUser)
		}
	}

	return usersList
}

// generateLokiIngressAuthSecret returns a secret for the loki ingress auth
func generateLokiIngressAuthSecret(lc loggedcluster.Interface, credentialsSecret *v1.Secret) (map[string][]byte, error) {
	users := make(map[string][]byte)
	// Loop on write users
	for _, writeUser := range listWriteUsers(credentialsSecret) {
		writePassword, err := loggingcredentials.GetPassword(lc, credentialsSecret, writeUser)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		password, err := bcrypt.GenerateFromPassword([]byte(writePassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		users[writeUser] = password
	}

	return users, nil
}
