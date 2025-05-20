package loggedcluster

import (
	appv1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// LoggingEnabledDefault defines if WCs logs are collected by default
const LoggingEnabledDefault = true

// Interface contains the definition of functions that can differ between each type of cluster
type Interface interface {
	client.Object
	HasLoggingEnabled() bool
	GetLoggingAgent() string
	SetLoggingAgent(string)
	GetKubeEventsLogger() string
	SetKubeEventsLogger(string)
	GetClusterName() string
	GetObject() client.Object
	UnwireLogging(currentApp appv1.App) *appv1.App
	WireLogging(currentApp appv1.App) *appv1.App
}
