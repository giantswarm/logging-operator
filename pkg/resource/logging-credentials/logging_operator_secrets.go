package loggingcredentials

import (
	"crypto/rand"
	"math/big"

	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
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
func LoggingCredentialsSecretMeta(lc loggedcluster.Interface) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      LoggingCredentialsName,
		Namespace: LoggingCredentialsNamespace,
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

// Generate a random 20-characters password
func genPassword() (string, error) {
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
// the observabilitybundle application to enable promtail.
func GenerateLoggingCredentialsBasicSecret(lc loggedcluster.Interface) *v1.Secret {
	secret := v1.Secret{
		ObjectMeta: LoggingCredentialsSecretMeta(lc),
		Data:       map[string][]byte{},
	}

	return &secret
}

func GetPassword(lc loggedcluster.Interface, credentialsSecret *v1.Secret, username string) (string, error) {
	var userYaml userCredentials

	userSecret, ok := credentialsSecret.Data[username]
	if !ok {
		return "", errors.New("Not found")
	}

	err := yaml.Unmarshal(userSecret, &userYaml)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Invalid user %s", username))
	}

	password := userYaml.Password

	return string(password), nil
}

// AddLoggingCredentials - Add credentials to LoggingCredentials secret if needed
func AddLoggingCredentials(lc loggedcluster.Interface, loggingCredentials *v1.Secret) (bool, error) {

	var secretUpdated bool = false

	// Always check credentials for "readuser"
	if _, ok := loggingCredentials.Data[common.ReadUser]; !ok {
		readUser := userCredentials{}

		password, err := genPassword()
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
	clusterName := lc.GetClusterName()
	if _, ok := loggingCredentials.Data[clusterName]; !ok {
		clusterUser := userCredentials{}

		password, err := genPassword()
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
func RemoveLoggingCredentials(lc loggedcluster.Interface, loggingCredentials *v1.Secret) bool {
	var secretUpdated bool = false

	// Check credentials for [clustername]
	clusterName := lc.GetClusterName()

	if _, ok := loggingCredentials.Data[clusterName]; ok {
		delete(loggingCredentials.Data, clusterName)
		secretUpdated = true
	}

	return secretUpdated
}
