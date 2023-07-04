package loggedcluster

import "sigs.k8s.io/controller-runtime/pkg/client"

// Interface contains the definition of functions that can differ between each type of cluster
type Interface interface {
	client.Object
	GetLoggingLabel() string
	GetAppsNamespace() string
	GetClusterName() string
	GetInstallationName() string
	AppConfigName(app string) string
	GetObject() client.Object
	GetObservabilityBundleConfigMap() string
}
