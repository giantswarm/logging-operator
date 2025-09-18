package tracingsecret

import (
	"golang.org/x/crypto/bcrypt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"
	"github.com/pkg/errors"
)

const (
	//#nosec G101
	tempoIngressAuthSecretName      = "tempo-ingress-auth"
	tempoIngressAuthSecretNamespace = "tempo"
	// DefaultReadOrgIDs - make sure to have at least 2 tenants, to prevent writing with this user
	DefaultReadOrgIDs = "giantswarm|default"
)

// tempoIngressAuthSecretMetadata returns metadata for the tempo ingress auth secret metadata
func tempoIngressAuthSecretMetadata() metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      tempoIngressAuthSecretName,
		Namespace: tempoIngressAuthSecretNamespace,
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

func tempoIngressAuthSecret() v1.Secret {
	return v1.Secret{
		ObjectMeta: tempoIngressAuthSecretMetadata(),
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

// generatetempoIngressAuthSecret returns a secret for the tempo ingress auth
func generatetempoIngressAuthSecret(cluster *capi.Cluster, credentialsSecret *v1.Secret) (map[string]string, error) {
	users := make(map[string]string)
	// Loop on write users
	for _, writeUser := range listWriteUsers(credentialsSecret) {
		writePassword, err := loggingcredentials.GetPassword(cluster, credentialsSecret, writeUser)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		password, err := bcrypt.GenerateFromPassword([]byte(writePassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		users[writeUser] = string(password)
	}

	return users, nil
}
