package predicates

import (
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

	if len(e.ObjectOld.GetLabels()) == 0 || len(e.ObjectNew.GetLabels()) == 0 ||
		e.ObjectOld.GetLabels()["app.kubernetes.io/name"] != "observability-bundle" ||
		e.ObjectNew.GetLabels()["app.kubernetes.io/name"] != "observability-bundle" {
		return false
	}

	return e.ObjectNew.GetResourceVersion() != e.ObjectOld.GetResourceVersion()
}
