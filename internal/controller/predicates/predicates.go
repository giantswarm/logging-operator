package predicates

import (
	"strings"

	"github.com/blang/semver"
	appv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ResourceVersionChangedPredicate implements a default update predicate function on resource version change.
type ObservabilityBundleAppVersionChangedPredicate struct {
	predicate.Funcs
}

// Update implements default UpdateEvent filter for validating resource version change.
func (ObservabilityBundleAppVersionChangedPredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil {
		return false
	}
	if e.ObjectNew == nil {
		return false
	}

	if !strings.Contains(e.ObjectOld.GetName(), "observability-bundle") || !strings.Contains(e.ObjectNew.GetName(), "observability-bundle") {
		return false
	}

	var oldApp, newApp *appv1alpha1.App
	var ok bool
	if oldApp, ok = e.ObjectOld.(*appv1alpha1.App); !ok {
		return false
	}
	if newApp, ok = e.ObjectNew.(*appv1alpha1.App); !ok {
		return false
	}

	oldAppVersion, err := semver.New(oldApp.Spec.Version)
	if err != nil {
		return false
	}
	newAppVersion, err := semver.New(newApp.Spec.Version)
	if err != nil {
		return false
	}
	breakingVersion := semver.MustParse("1.0.0")
	return oldAppVersion.LT(breakingVersion) && newAppVersion.GTE(breakingVersion)
}
