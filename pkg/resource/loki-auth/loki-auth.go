package lokiauth

import (
	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	lokiauthSecretName      = "loki-multi-tenant-proxy-auth-config"
	lokiauthSecretNamespace = "loki"
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

// GenerateLokiAuthSecret returns a secret for
// the Loki-multi-tenant-proxy auth config
func GenerateLokiAuthSecret(lc loggedcluster.Interface, credentialsSecret *v1.Secret) (v1.Secret, error) {

	readUser, err := loggingcredentials.GetLogin(lc, credentialsSecret, "read")
	if err != nil {
		return v1.Secret{}, errors.WithStack(err)
	}

	readPassword, err := loggingcredentials.GetPass(lc, credentialsSecret, "read")
	if err != nil {
		return v1.Secret{}, errors.WithStack(err)
	}

	writeUser, err := loggingcredentials.GetLogin(lc, credentialsSecret, "write")
	if err != nil {
		return v1.Secret{}, errors.WithStack(err)
	}

	writePassword, err := loggingcredentials.GetPass(lc, credentialsSecret, "write")
	if err != nil {
		return v1.Secret{}, errors.WithStack(err)
	}

	values := Values{
		Users: []user{
			{
				Username: readUser,
				Password: readPassword,
				// make sure to have at least 2 tenants to prevent writing with this user
				Orgid: "giantswarm|default",
			},
			{
				Username: writeUser,
				Password: writePassword,
				// on the write path the tenant will be given by the sender (promtail)
				Orgid: "none",
			},
		},
	}

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
