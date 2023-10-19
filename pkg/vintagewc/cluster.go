package vintagewc

import (
	"fmt"
	"strconv"

	appv1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"golang.org/x/mod/semver"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/logging-operator/pkg/key"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

type Object struct {
	client.Object
	Options loggedcluster.Options
}

func (o Object) HasLoggingEnabled() bool {
	labels := o.Object.GetLabels()

	// If logging is disabled at the installation level, we return false
	if !o.Options.EnableLoggingFlag {
		return false
	}

	// Promtail only works starting with AWS version 19.1.0
	clusterRelease := labels["release.giantswarm.io/version"]
	// semver versions must be "vMAJOR[.MINOR[.PATCH[-PRERELEASE][+BUILD]]]"
	clusterReleaseSemver := fmt.Sprintf("v%s", clusterRelease)
	if semver.Compare(clusterReleaseSemver, "v19.1.0") == -1 {
		return false
	}

	loggingLabelValue, ok := labels[key.LoggingLabel]
	if !ok {
		// This is what we will have to change when we enable logging on all WCs >= 19.1.0
		return false
	}

	loggingEnabled, err := strconv.ParseBool(loggingLabelValue)
	if err != nil {
		return false
	}
	return loggingEnabled
}

func (o Object) GetAppsNamespace() string {
	return o.Object.GetName()
}

func (o Object) AppConfigName(app string) string {
	if app == "observability-bundle" {
		return o.GetObject().GetName() + "-" + app
	} else {
		return app
	}
}

func (o Object) ObservabilityBundleConfigLabelName(config string) string {
	return config
}

func (o Object) GetClusterName() string {
	return o.Object.GetName()
}

func (o Object) GetInstallationName() string {
	return o.Options.InstallationName
}

func (o Object) GetEnableLoggingFlag() bool {
	return o.Options.EnableLoggingFlag
}

func (o Object) GetObject() client.Object {
	return o.Object
}

// on vintage WC, use extraconfig
func (o Object) GetObservabilityBundleConfigMap() string {
	return "observability-bundle-logging-extraconfig"
}

func (o Object) UnwirePromtail(currentApp appv1.App) *appv1.App {
	// cluster-operator is taking care of the unwiring, nothing to do here
	return currentApp.DeepCopy()
}

func (o Object) WirePromtail(currentApp appv1.App) *appv1.App {
	// cluster-operator is taking care of the wiring, nothing to do here
	return currentApp.DeepCopy()
}
