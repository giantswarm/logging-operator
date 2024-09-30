package podlogs

import (
	podlogsv1alpha2 "github.com/giantswarm/logging-operator/pkg/resource/podlogs/apis/monitoring/v1alpha2"
)

// PodLogsGetter is a helper to create PodLogs objects
// using controllerutil.CreateOrUpdate
// One can pass the object reference returned by GetWithMetaOnly to CreateOrUpdate
// Then update the PodLogs spec with GetSpec.
type PodLogsGetter struct {
	podlogsv1alpha2.PodLogs
}

// GetWithMetaOnly returns a PodLogs object only with ObjectMeta populated.
func (p *PodLogsGetter) GetWithMetaOnly() *podlogsv1alpha2.PodLogs {
	pl := podlogsv1alpha2.PodLogs{
		ObjectMeta: p.ObjectMeta,
	}

	return &pl
}

// GetSpec returns the PodLogs spec.
func (p *PodLogsGetter) GetSpec() podlogsv1alpha2.PodLogsSpec {
	return p.Spec
}
