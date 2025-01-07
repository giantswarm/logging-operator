package proxyauth

import (
	"fmt"
	"strings"

	"github.com/giantswarm/grafana-multi-tenant-proxy/pkg/config"
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
	proxyauthSecretName      = "grafana-multi-tenant-proxy-auth-config"
	proxyauthSecretNamespace = "monitoring"
	// DefaultReadOrgIDs - make sure to have at least 2 tenants, to prevent writing with this user
	DefaultReadOrgIDs = "giantswarm|default"
)

// ProxyConfigSecretMeta returns metadata for the grafana-multi-tenant-proxy secret
func proxyAuthSecretMeta(lc loggedcluster.Interface) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      proxyauthSecretName,
		Namespace: proxyauthSecretNamespace,
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

// GenerateProxyAuthSecret returns a secret for
// the grafana-multi-tenant-proxy auth config
func GenerateProxyAuthSecret(lc loggedcluster.Interface, credentialsSecret *v1.Secret) (v1.Secret, error) {
	// Init empty users structure
	authCfg := config.AuthenticationConfig{
		Users: []config.User{},
	}
	// Prepare read user's orgid with default values
	readOrgid := DefaultReadOrgIDs

	// Loop on write users
	for _, writeUser := range listWriteUsers(credentialsSecret) {
		writePassword, err := loggingcredentials.GetPassword(lc, credentialsSecret, writeUser)
		if err != nil {
			return v1.Secret{}, errors.WithStack(err)
		}

		tenant := writeUser
		if lc.IsCAPI() {
			tenant = lc.GetTenant()
		}

		authCfg.Users = append(authCfg.Users, config.User{
			Username: writeUser,
			Password: writePassword,
			// we set the default tenant even though it may be given by the sender
			// depending of grafana-multi-teant-proxy config
			OrgID: tenant,
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
	authCfg.Users = append(authCfg.Users, config.User{
		Username: readUser,
		Password: readPassword,
		OrgID:    readOrgid,
	})

	v, err := yaml.Marshal(authCfg)
	if err != nil {
		return v1.Secret{}, errors.WithStack(err)
	}
	secret := secret()
	secret.Data["authn.yaml"] = []byte(v)

	return secret, nil
}

func secret() v1.Secret {
	return v1.Secret{
		ObjectMeta: proxyAuthSecretMeta(nil),
		Data:       map[string][]byte{},
	}
}
