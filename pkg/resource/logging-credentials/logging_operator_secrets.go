package loggingcredentials

import (
	"crypto/rand"
	"math/big"

	"fmt"

	"github.com/pkg/errors"

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

func GetLogin(lc loggedcluster.Interface, credentialsSecret *v1.Secret, user string) (string, error) {

	login, ok := credentialsSecret.Data[fmt.Sprintf("%suser", user)]

	if !ok {
		return "", errors.New("Not found")
	}
	return string(login), nil
}

func GetPass(lc loggedcluster.Interface, credentialsSecret *v1.Secret, user string) (string, error) {
	pass, ok := credentialsSecret.Data[fmt.Sprintf("%spassword", user)]

	if !ok {
		return "", errors.New("Not found")
	}
	return string(pass), nil
}

// AddLoggingCredentials - Add credentials to LoggingCredentials secret if needed
func AddLoggingCredentials(lc loggedcluster.Interface, loggingCredentials *v1.Secret) bool {

	var secretUpdated bool = false

	// Always check credentials for "readuser"
	if _, ok := loggingCredentials.Data["readuser"]; !ok {
		loggingCredentials.Data["readuser"] = []byte("read")
		secretUpdated = true
	}
	if _, ok := loggingCredentials.Data["readpassword"]; !ok {
		password, err := genPassword()
		if err != nil {
			return false
		}
		loggingCredentials.Data["readpassword"] = []byte(password)
		secretUpdated = true
	}

	// Check credentials for [clustername]
	clusterName := lc.GetClusterName()
	if _, ok := loggingCredentials.Data[fmt.Sprintf("%suser", clusterName)]; !ok {
		loggingCredentials.Data[fmt.Sprintf("%suser", clusterName)] = []byte(clusterName)
		secretUpdated = true
	}
	if _, ok := loggingCredentials.Data[fmt.Sprintf("%spassword", clusterName)]; !ok {
		password, err := genPassword()
		if err != nil {
			return false
		}
		loggingCredentials.Data[fmt.Sprintf("%spassword", clusterName)] = []byte(password)
		secretUpdated = true
	}

	return secretUpdated
}

// RemoveLoggingCredentials - Remove credentials from LoggingCredentials secret
func RemoveLoggingCredentials(lc loggedcluster.Interface, loggingCredentials *v1.Secret) bool {
	var secretUpdated bool = false

	// Check credentials for [clustername]
	clusterName := lc.GetClusterName()
	credsUsername := fmt.Sprintf("%suser", clusterName)
	credsPassword := fmt.Sprintf("%spassword", clusterName)

	if _, ok := loggingCredentials.Data[credsUsername]; ok {
		delete(loggingCredentials.Data, credsUsername)
		secretUpdated = true
	}

	if _, ok := loggingCredentials.Data[credsPassword]; ok {
		delete(loggingCredentials.Data, credsPassword)
		secretUpdated = true
	}
	return secretUpdated
}
