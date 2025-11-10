package agentstoggle

import (
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/logging-operator/pkg/common"
)

type values struct {
	Apps map[string]app `yaml:"apps" json:"apps"`
}

type app struct {
	Enabled   bool   `yaml:"enabled" json:"enabled"`
	Namespace string `yaml:"namespace,omitempty" json:"namespace,omitempty"`
}

// generateObservabilityBundleConfig returns a configmap for
// the observabilitybundle application to enable logging agents and events-loggers.
func generateObservabilityBundleConfig() (string, error) {
	values := values{
		Apps: map[string]app{
			common.AlloyLogsObservabilityBundleAppName: {
				Enabled: true,
			},
			common.AlloyEventsObservabilityBundleAppName: {
				Enabled: true,
			},
		},
	}

	v, err := yaml.Marshal(values)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return string(v), nil
}
