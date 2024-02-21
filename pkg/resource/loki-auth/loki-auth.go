package lokiauth

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"
)

const (
	//#nosec G101
	lokiauthSecretName      = "loki-multi-tenant-proxy-auth-config"
	lokiauthSecretNamespace = "loki"
	// DefaultReadOrgIDs - make sure to have at least 2 tenants, to prevent writing with this user
	DefaultReadOrgIDs = "giantswarm|default"
)

type Values struct {
	Users []user `yaml:"users" json:"users"`
}

type user struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	Orgid    string `yaml:"orgid" json:"orgid"`
}

// LokiConfigSecretMeta returns metadata for the Loki-multi-tenant-proxy secret
func LokiAuthSecretMeta(lc loggedcluster.Interface) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      lokiauthSecretName,
		Namespace: lokiauthSecretNamespace,
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

// listWriteUsers returns a map of users found in a credentialsSecret
func listWriteUsers(credentialsSecret *v1.Secret) []string {
	var usersList []string
	for myUser := range credentialsSecret.Data {
		// bypass read user
		// bypass old creds (xxxuser and xxxpassword)

		if !strings.HasSuffix(myUser, "user") && !strings.HasSuffix(myUser, "password") && myUser != common.ReadUser {
			usersList = append(usersList, myUser)
		}
	}

	return usersList
}

// GenerateLokiAuthSecret returns a secret for
// the Loki-multi-tenant-proxy auth config
func GenerateLokiAuthSecret(lc loggedcluster.Interface, credentialsSecret *v1.Secret) (v1.Secret, error) {
	// Init empty users structure
	values := Values{
		Users: []user{},
	}
	// Prepare read user's orgid with default values
	readOrgid := DefaultReadOrgIDs

	// Loop on write users
	for _, writeUser := range listWriteUsers(credentialsSecret) {

		writePassword, err := loggingcredentials.GetPassword(lc, credentialsSecret, writeUser)
		if err != nil {
			return v1.Secret{}, errors.WithStack(err)
		}

		values.Users = append(values.Users, user{
			Username: writeUser,
			Password: writePassword,
			// we set the tenant even though it may be given by the sender (promtail)
			// depending of loki-multi-teant-proxy config
			Orgid: writeUser,
		})

		// Add write user to allowed tenants for read user
		readOrgid = fmt.Sprintf("%s|%s", readOrgid, writeUser)
	}

	// Create read user
	readUser := common.ReadUser

	readPassword, err := loggingcredentials.GetPassword(lc, credentialsSecret, readUser)
	if err != nil {
		return v1.Secret{}, errors.WithStack(err)
	}
	values.Users = append(values.Users, user{
		Username: readUser,
		Password: readPassword,
		Orgid:    readOrgid,
	})

	v, err := yaml.Marshal(values)
	if err != nil {
		return v1.Secret{}, errors.WithStack(err)
	}

	secret := v1.Secret{
		ObjectMeta: LokiAuthSecretMeta(lc),
		Data: map[string][]byte{
			"authn.yaml": []byte(v),
		},
	}

	return secret, nil
}
