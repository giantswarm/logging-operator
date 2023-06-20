package loggedcluster

import "sigs.k8s.io/controller-runtime/pkg/client"

type Interface interface {
	client.Object
	GetLoggingLabel() string
	GetAppsNamespace() string
	GetAppName(app string) string
}
