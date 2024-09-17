package grafanadatasource

import (
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"
)

const (
	lokiURL                   = "http://grafana-multi-tenant-proxy.monitoring.svc"
	datasourceName            = "Loki"
	datasourceSecretName      = "loki-datasource"
	datasourceSecretNamespace = "monitoring"
	datasourceFileName        = "loki-datasource.yaml"
)

type Values struct {
	ApiVersion  int          `yaml:"apiVersion" json:"apiVersion"`
	Datasources []datasource `yaml:"datasources" json:"datasources"`
}

type datasource struct {
	Access         string         `yaml:"access" json:"access"`
	Editable       bool           `yaml:"editable" json:"editable"`
	BasicAuth      bool           `yaml:"basicAuth" json:"basicAuth"`
	BasicAuthUser  string         `yaml:"basicAuthUser" json:"basicAuthUser"`
	JsonData       jsonData       `yaml:"jsonData" json:"jsonData"`
	Name           string         `yaml:"name" json:"name"`
	Type           string         `yaml:"type" json:"type"`
	Url            string         `yaml:"url" json:"url"`
	SecureJsonData secureJsonData `yaml:"secureJsonData" json:"secureJsonData"`
}

type jsonData struct {
	ManageAlerts bool `yaml:"manageAlerts" json:"manageAlerts"`
}

type secureJsonData struct {
	BasicAuthPassword string `yaml:"basicAuthPassword" json:"basicAuthPassword"`
}

// DatasourceSecretMeta returns metadata for the observability bundle extra values configmap.
func DatasourceSecretMeta(lc loggedcluster.Interface) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      datasourceSecretName,
		Namespace: datasourceSecretNamespace,
		Labels: map[string]string{
			// This label is used to detect datasources
			"app.giantswarm.io/kind": "datasource",
		},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

// GenerateDatasourceSecret returns a secret for
// the Loki datasource for Grafana
func GenerateDatasourceSecret(lc loggedcluster.Interface, credentialsSecret *v1.Secret) (v1.Secret, error) {
	user := common.ReadUser

	password, err := loggingcredentials.GetPassword(lc, credentialsSecret, user)
	if err != nil {
		return v1.Secret{}, errors.WithStack(err)
	}

	values := Values{
		ApiVersion: 1,
		Datasources: []datasource{
			{
				Access:        "proxy",
				Editable:      false,
				BasicAuth:     true,
				BasicAuthUser: user,
				JsonData: jsonData{
					ManageAlerts: false,
				},
				Name: datasourceName,
				Type: "loki",
				Url:  lokiURL,
				SecureJsonData: secureJsonData{
					BasicAuthPassword: password,
				},
			},
		},
	}

	v, err := yaml.Marshal(values)
	if err != nil {
		return v1.Secret{}, errors.WithStack(err)
	}

	secret := v1.Secret{
		ObjectMeta: DatasourceSecretMeta(lc),
		Data: map[string][]byte{
			datasourceFileName: []byte(v),
		},
	}

	return secret, nil
}
