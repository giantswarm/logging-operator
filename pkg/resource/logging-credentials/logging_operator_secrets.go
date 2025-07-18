package loggingcredentials

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/logging-operator/pkg/common"
)

const (
	//#nosec G101
	LoggingCredentialsName      = "logging-credentials"
	LoggingCredentialsNamespace = "monitoring"
)

type userCredentials struct {
	Password string `yaml:"password" json:"password"`
}

// LoggingCredentialsSecretMeta returns metadata for the logging-operator credentials secret.
func LoggingCredentialsSecretMeta() metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      LoggingCredentialsName,
		Namespace: LoggingCredentialsNamespace,
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

// GenerateObservabilityBundleConfigMap returns a configmap for
// the observabilitybundle application to enable logging.
func GenerateLoggingCredentialsBasicSecret(cluster *capi.Cluster) *v1.Secret {
	secret := v1.Secret{
		ObjectMeta: LoggingCredentialsSecretMeta(),
		Data:       map[string][]byte{},
	}

	return &secret
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

// AddLoggingCredentials - Add credentials to LoggingCredentials secret if needed
func AddLoggingCredentials(cluster *capi.Cluster, loggingCredentials *v1.Secret) (bool, error) {
	var secretUpdated = false

	// Always check credentials for "readuser"
	if _, ok := loggingCredentials.Data[common.ReadUser]; !ok {
		readUser := userCredentials{}

		password, err := generatePassword()
		if err != nil {
			return false, errors.New("Failed generating read password")
		}

		readUser.Password = password

		v, err := yaml.Marshal(readUser)
		if err != nil {
			return false, errors.New("Failed creating read user")
		}

		loggingCredentials.Data[common.ReadUser] = []byte(v)
		secretUpdated = true
	}

	// Check credentials for [clustername]
	clusterName := cluster.GetName()
	if _, ok := loggingCredentials.Data[clusterName]; !ok {
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

		loggingCredentials.Data[clusterName] = []byte(v)
		secretUpdated = true
	}

	return secretUpdated, nil
}

// RemoveLoggingCredentials - Remove credentials from LoggingCredentials secret
func RemoveLoggingCredentials(cluster *capi.Cluster, loggingCredentials *v1.Secret) bool {
	var secretUpdated = false

	// Check credentials for [clustername]
	clusterName := cluster.GetName()

	if _, ok := loggingCredentials.Data[clusterName]; ok {
		delete(loggingCredentials.Data, clusterName)
		secretUpdated = true
	}

	return secretUpdated
}
