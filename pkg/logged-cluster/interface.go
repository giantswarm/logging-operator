package loggedcluster

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Interface contains the definition of functions that can differ between each type of cluster
type Interface interface {
	client.Object
	GetLoggingAgent() string
	SetLoggingAgent(string)
	GetKubeEventsLogger() string
	SetKubeEventsLogger(string)
}
