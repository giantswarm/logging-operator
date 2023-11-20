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

// /// Promtail values config structure
type values struct {
	Promtail promtail `yaml:"promtail" json:"promtail"`
}

type promtail struct {
	ExtraArgs         []string                   `yaml:"extraArgs" json:"extraArgs"`
	ExtraEnv          []promtailExtraEnv         `yaml:"extraEnv" json:"extraEnv"`
	Config            promtailConfig             `yaml:"config" json:"config"`
	ExtraVolumes      []promtailExtraVolume      `yaml:"extraVolumes" json:"extraVolumes"`
	ExtraVolumeMounts []promtailExtraVolumeMount `yaml:"extraVolumeMounts" json:"extraVolumeMounts"`
}

type promtailExtraEnvValuefrom struct {
	FieldRef promtailExtraEnvFieldref `yaml:"fieldRef" json:"fieldRef"`
}

type promtailExtraEnvFieldref struct {
	FieldPath string `yaml:"fieldPath" json:"fieldPath"`
}

type promtailExtraEnv struct {
	Name      string                    `yaml:"name" json:"name"`
	ValueFrom promtailExtraEnvValuefrom `yaml:"valueFrom" json:"valueFrom"`
}

type promtailConfigSnippets struct {
	ExtraScrapeConfigs  string               `yaml:"extraScrapeConfigs" json:"extraScrapeConfigs"`
	ExtraRelabelConfigs []extraRelabelConfig `yaml:"extraRelabelConfigs" json:"extraRelabelConfigs"`
	AddScrapeJobLabel   bool                 `yaml:"addScrapeJobLabel" json:"addScrapeJobLabel"`
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

type extraRelabelConfig struct {
	SourceLabels []string `yaml:"source_labels" json:"source_labels"`
	Action       string   `yaml:"action" json:"action"`
	Regex        string   `yaml:"regex" json:"regex"`
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
	// Scrape logs from kube-system and giantswarm namespaces only for WC clusters
	var extraRelabelConfigs []extraRelabelConfig
	if common.IsWorkloadCluster(lc) {
		extraNamespaces := extraRelabelConfig{
			SourceLabels: []string{
				"__meta_kubernetes_namespace",
			},
			Action: "keep",
			Regex:  "kube-system|giantswarm",
		}
		extraRelabelConfigs = append(extraRelabelConfigs, extraNamespaces)
	}

	values := values{
		Promtail: promtail{
			ExtraArgs: []string{
				"-log-config-reverse-order",
				"-config.expand-env=true",
			},
			ExtraEnv: []promtailExtraEnv{
				{
					Name: "NODENAME",
					ValueFrom: promtailExtraEnvValuefrom{
						FieldRef: promtailExtraEnvFieldref{
							FieldPath: "spec.nodeName",
						},
					},
				},
			},
			Config: promtailConfig{
				Snippets: promtailConfigSnippets{
					ExtraScrapeConfigs: `# this one includes also system logs reported by systemd-journald
- job_name: systemd_journal_run
  journal:
    path: /run/log/journal
    max_age: 12h
    json: true
    labels:
      scrape_job: system-logs
  relabel_configs:
  - source_labels: ['__journal__systemd_unit']
    target_label: 'systemd_unit'
  - source_labels: ['__journal__hostname']
    target_label: 'node_name'
- job_name: systemd_journal_var
  journal:
    path: /var/log/journal
    max_age: 12h
    json: true
    labels:
      scrape_job: system-logs
  relabel_configs:
  - source_labels: ['__journal__systemd_unit']
    target_label: 'systemd_unit'
  - source_labels: ['__journal__hostname']
    target_label: 'node_name'
- job_name: kubernetes-audit
  static_configs:
  - targets:
    - localhost
    labels:
      scrape_job: audit-logs
      __path__: /var/log/apiserver/*.log
      node_name: ${NODENAME:-unknown}
  pipeline_stages:
  - json:
      expressions:
        objectRef: objectRef
        level:     level
        stage:     stage
        verb:      verb
  - json:
      expressions:
        resource: resource
        namespace: namespace
        name: name
      source: objectRef
  - labels:
      level:
      stage:
      verb:
      resource:
      namespace:
      name:
`,
					ExtraRelabelConfigs: extraRelabelConfigs,
					AddScrapeJobLabel:   true,
				},
			},
			ExtraVolumes: []promtailExtraVolume{
				{
					Name: "journal-run",
					HostPath: promtailExtraVolumeHostpath{
						Path: "/run/log/journal/",
					},
				},
				{
					Name: "journal-var",
					HostPath: promtailExtraVolumeHostpath{
						Path: "/var/log/journal/",
					},
				},
				{
					Name: "apiserver-logs",
					HostPath: promtailExtraVolumeHostpath{
						Path: "/var/log/apiserver/",
					},
				},
			},
			ExtraVolumeMounts: []promtailExtraVolumeMount{
				{
					Name:      "journal-run",
					MountPath: "/run/log/journal/",
					ReadOnly:  true,
				},
				{
					Name:      "journal-var",
					MountPath: "/var/log/journal/",
					ReadOnly:  true,
				},
				{
					Name:      "apiserver-logs",
					MountPath: "/var/log/apiserver/",
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
