package loggingsecret

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

const (
	loggingClientSecretName = "logging-secret"
)

// SecretMeta returns metadata for the logging-secret
func SecretMeta(lc loggedcluster.Interface) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      getLoggingSecretName(lc),
		Namespace: lc.GetAppsNamespace(),
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

func getLoggingSecretName(lc loggedcluster.Interface) string {
	return fmt.Sprintf("%s-%s", lc.GetClusterName(), loggingClientSecretName)
}
