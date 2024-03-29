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
	IsInsecureCA() bool
	GetAppsNamespace() string
	GetEnableLoggingFlag() bool
	GetClusterName() string
	GetInstallationName() string
	AppConfigName(app string) string
	ObservabilityBundleConfigLabelName(config string) string
	GetObject() client.Object
	GetObservabilityBundleConfigMap() string
	UnwirePromtail(currentApp appv1.App) *appv1.App
	WirePromtail(currentApp appv1.App) *appv1.App
}
