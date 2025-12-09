package credentials

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/core/v1beta1"

	"github.com/giantswarm/logging-operator/pkg/common"
)

const (
	LoggingCredentialsName      = "logging-credentials" // #nosec G101
	LoggingCredentialsNamespace = "monitoring"
	TracingCredentialsName      = "tracing-credentials" // #nosec G101
	TracingCredentialsNamespace = "monitoring"
)

type userCredentials struct {
	Password string `yaml:"password" json:"password"`
}

// credentialsSecretMeta returns metadata for the logging-operator credentials secret.
func CredentialsSecretMeta(name string, namespace string) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

// Generate a random 20-characters password
func generatePassword() (string, error) {
	const length = 20

	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")

	pass := make([]rune, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", errors.WithStack(err)
		}
		pass[i] = chars[num.Int64()]
	}
	return string(pass), nil
}

func GenerateLoggingCredentialsBasicSecret(cluster *capi.Cluster) *v1.Secret {
	loggingSecret := v1.Secret{
		ObjectMeta: CredentialsSecretMeta(LoggingCredentialsName, LoggingCredentialsNamespace),
		Data:       map[string][]byte{},
	}

	return &loggingSecret
}

func GenerateTracingCredentialsBasicSecret(cluster *capi.Cluster) *v1.Secret {
	tracingSecret := v1.Secret{
		ObjectMeta: CredentialsSecretMeta(TracingCredentialsName, TracingCredentialsNamespace),
		Data:       map[string][]byte{},
	}

	return &tracingSecret
}

func GetPassword(cluster *capi.Cluster, credentialsSecret *v1.Secret, username string) (string, error) {
	var userYaml userCredentials

	userSecret, ok := credentialsSecret.Data[username]
	if !ok {
		return "", errors.New("Not found")
	}

	err := yaml.Unmarshal(userSecret, &userYaml)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Invalid user %s", username))
	}

	return string(userYaml.Password), nil
}

// AddCredentials - Add credentials to secret if needed
func AddCredentials(cluster *capi.Cluster, credentials *v1.Secret) (bool, error) {
	var secretUpdated = false

	// Check credentials for [clustername]
	clusterName := cluster.GetName()
	if _, ok := credentials.Data[clusterName]; !ok {
		clusterUser := userCredentials{}

		password, err := generatePassword()
		if err != nil {
			return false, errors.New("Failed generating write password")
		}

		clusterUser.Password = password

		v, err := yaml.Marshal(clusterUser)
		if err != nil {
			return false, errors.New("Failed creating write user")
		}

		credentials.Data[clusterName] = []byte(v)
		secretUpdated = true
	}

	return secretUpdated, nil
}

// RemoveCredentials - Remove credentials from credentials secret
func RemoveCredentials(cluster *capi.Cluster, credentials *v1.Secret) bool {
	var secretUpdated = false

	// Check credentials for [clustername]
	clusterName := cluster.GetName()

	if _, ok := credentials.Data[clusterName]; ok {
		delete(credentials.Data, clusterName)
		secretUpdated = true
	}

	return secretUpdated
}
