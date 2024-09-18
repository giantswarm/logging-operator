package loggingconfig

import (
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
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
	PipelineStages      []map[interface{}]interface{} `yaml:"pipelineStages" json:"pipelineStages"`
	ExtraScrapeConfigs  string                        `yaml:"extraScrapeConfigs" json:"extraScrapeConfigs"`
	ExtraRelabelConfigs []extraRelabelConfig          `yaml:"extraRelabelConfigs" json:"extraRelabelConfigs"`
	AddScrapeJobLabel   bool                          `yaml:"addScrapeJobLabel" json:"addScrapeJobLabel"`
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

// GeneratePromtailLoggingConfig returns a configmap for
// the logging extra-config
func GeneratePromtailLoggingConfig(lc loggedcluster.Interface) (string, error) {
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
					PipelineStages: []map[interface{}]interface{}{
						{
							"cri": map[interface{}]interface{}{},
						},
						{
							"structured_metadata": map[interface{}]interface{}{
								"filename": nil,
								"stream":   nil,
							},
						},
						{
							"labeldrop": map[interface{}]interface{}{
								"filename": nil,
								"stream":   nil,
							},
						},
					},
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
    target_label: '__tmp_systemd_unit'
  - source_labels:
    - __journal__systemd_unit
    - __journal_syslog_identifier
    separator: ;
    regex: ';(.+)'
    replacement: $1
    target_label: '__tmp_systemd_unit'
  - source_labels: ['__tmp_systemd_unit']
    target_label: 'systemd_unit'
  - source_labels: ['__journal__hostname']
    target_label: 'node'
  pipeline_stages:
  - json:
      expressions:
        SYSLOG_IDENTIFIER: SYSLOG_IDENTIFIER
  - drop:
      source: SYSLOG_IDENTIFIER
      value: audit
- job_name: systemd_journal_var
  journal:
    path: /var/log/journal
    max_age: 12h
    json: true
    labels:
      scrape_job: system-logs
  relabel_configs:
  - source_labels: ['__journal__systemd_unit']
    target_label: '__tmp_systemd_unit'
  - source_labels:
    - __journal__systemd_unit
    - __journal_syslog_identifier
    separator: ;
    regex: ';(.+)'
    replacement: $1
    target_label: '__tmp_systemd_unit'
  - source_labels: ['__tmp_systemd_unit']
    target_label: 'systemd_unit'
  - source_labels: ['__journal__hostname']
    target_label: 'node'
  pipeline_stages:
  - json:
      expressions:
        SYSLOG_IDENTIFIER: SYSLOG_IDENTIFIER
  - drop:
      source: SYSLOG_IDENTIFIER
      value: audit
- job_name: kubernetes-audit
  static_configs:
  - targets:
    - localhost
    labels:
      scrape_job: audit-logs
      __path__: /var/log/apiserver/audit.log
      node: ${NODENAME:-unknown}
  pipeline_stages:
  - json:
      expressions:
        objectRef: objectRef
  - json:
      expressions:
        resource: resource
        namespace: namespace
      source: objectRef
  - structured_metadata:
      resource:
      filename:
  - labeldrop:
      filename
  - labels:
      namespace:
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
		return "", errors.WithStack(err)
	}

	return string(v), nil
}
