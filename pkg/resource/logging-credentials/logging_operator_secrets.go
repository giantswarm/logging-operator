package loggingcredentials

import (
	"math/rand"
	"strings"
	"time"

	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LoggingCredentialsSecretMeta returns metadata for the logging-operator credentials secret.
func LoggingCredentialsSecretMeta(lc loggedcluster.Interface) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      lc.AppConfigName("logging-credentials"),
		Namespace: "monitoring",
	}
}

// Generate a random 20-characters password
func genPassword() string {
	const length = 20

	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789" +
		"~=+%^*/()[]{}/!@#$?|")

	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])

	}
	str := b.String() // E.g. "ExcbsVQs"
	return str
}

// GenerateObservabilityBundleConfigMap returns a configmap for
// the observabilitybundle application to enable promtail.
func GenerateLoggingCredentialsBasicSecret(lc loggedcluster.Interface) (v1.Secret, error) {

	secret := v1.Secret{
		ObjectMeta: LoggingCredentialsSecretMeta(lc),
		Data:       map[string][]byte{},
	}

	return secret, nil
}

// Update a LoggingCredentials secret if needed
func UpdateLoggingCredentials(loggingCredentials v1.Secret) bool {

	var secretUpdated bool = false

	if _, ok := loggingCredentials.Data["readuser"]; !ok {
		loggingCredentials.Data["readuser"] = []byte("read")
		secretUpdated = true
	}
	if _, ok := loggingCredentials.Data["readpassword"]; !ok {
		loggingCredentials.Data["readpassword"] = []byte(genPassword())
		secretUpdated = true
	}
	if _, ok := loggingCredentials.Data["writeuser"]; !ok {
		loggingCredentials.Data["writeuser"] = []byte("write")
		secretUpdated = true
	}
	if _, ok := loggingCredentials.Data["writepassword"]; !ok {
		loggingCredentials.Data["writepassword"] = []byte(genPassword())
		secretUpdated = true
	}

	return secretUpdated
}
