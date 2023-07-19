package promtailconfig

import (
	"fmt"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

const (
	promtailConfigName = "logging-config"
)

///// Promtail values config structure

type values struct {
	Promtail promtail `yaml:"promtail" json:"promtail"`
}

type promtail struct {
	ExtraArgs         []string                   `yaml:"extraArgs" json:"extraArgs"`
	Config            promtailConfig             `yaml:"config" json:"config"`
	ExtraVolumes      []promtailExtraVolume      `yaml:"extraVolumes" json:"extraVolumes"`
	ExtraVolumeMounts []promtailExtraVolumeMount `yaml:"extraVolumeMounts" json:"extraVolumeMounts"`
}

type promtailConfigSnippets struct {
	ExtraScrapeConfigs string `yaml:"extraScrapeConfigs" json:"extraScrapeConfigs"`
}

type promtailConfig struct {
	Snippets promtailConfigSnippets `yaml:"snippets" json:"snippets"`
}

type promtailExtraVolume struct {
	Name     string                      `yaml:"name" json:"name"`
	HostPath promtailExtraVolumeHostpath `yaml:"hostPath" json:"hostPath"`
}

type promtailExtraVolumeHostpath struct {
	Path string `yaml:"path" json:"path"`
}

type promtailExtraVolumeMount struct {
	Name      string `yaml:"name" json:"name"`
	MountPath string `yaml:"mountPath" json:"mountPath"`
	ReadOnly  bool   `yaml:"readOnly" json:"readOnly"`
}

// ConfigMeta returns metadata for the promtail-config
func ConfigMeta(lc loggedcluster.Interface) metav1.ObjectMeta {
	metadata := metav1.ObjectMeta{
		Name:      fmt.Sprintf("%s-%s", lc.GetClusterName(), promtailConfigName),
		Namespace: lc.GetAppsNamespace(),
		Labels:    map[string]string{},
	}

	common.AddCommonLabels(metadata.Labels)
	return metadata
}

// GeneratePromtailConfig returns a configmap for
// the promtail extra-config
func GeneratePromtailConfig(lc loggedcluster.Interface) (v1.ConfigMap, error) {

	values := values{
		Promtail: promtail{
			ExtraArgs: []string{
				"-log-config-reverse-order",
			},
			Config: promtailConfig{
				Snippets: promtailConfigSnippets{
					ExtraScrapeConfigs: `# this one includes also system logs reported by systemd-journald
- job_name: systemd_journal
  journal:
    path: /run/log/journal
    max_age: 12h
    json: true
  relabel_configs:
    - source_labels: ['__journal__systemd_unit']
      target_label: 'systemd_unit'
    - source_labels: ['__journal__hostname']
      target_label: 'hostname'`,
				},
			},
			ExtraVolumes: []promtailExtraVolume{
				{
					Name: "journal",
					HostPath: promtailExtraVolumeHostpath{
						Path: "/run/log/journal/",
					},
				},
			},
			ExtraVolumeMounts: []promtailExtraVolumeMount{
				{
					Name:      "journal",
					MountPath: "/run/log/journal/",
					ReadOnly:  true,
				},
			},
		},
	}

	v, err := yaml.Marshal(values)
	if err != nil {
		return v1.ConfigMap{}, errors.WithStack(err)
	}

	configmap := v1.ConfigMap{
		ObjectMeta: ConfigMeta(lc),
		Data: map[string]string{
			"values": string(v),
		},
	}

	return configmap, nil
}
