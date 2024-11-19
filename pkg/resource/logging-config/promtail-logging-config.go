package loggingconfig

import (
	"bytes"
	"text/template"

	_ "embed"

	"github.com/Masterminds/sprig/v3"
	"github.com/blang/semver"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

var (
	//go:embed promtail/logging-config.promtail.yaml.template
	promtailLoggingConfig         string
	promtailLoggingConfigTemplate *template.Template

	supportsStructuredMetadata = semver.MustParse("1.0.0")
)

func init() {
	promtailLoggingConfigTemplate = template.Must(template.New("logging-config.promtail.yaml").Funcs(sprig.FuncMap()).Parse(promtailLoggingConfig))
}

// GeneratePromtailLoggingConfig returns a configmap for
// the logging extra-config
func GeneratePromtailLoggingConfig(lc loggedcluster.Interface, observabilityBundleVersion semver.Version) (string, error) {
	var values bytes.Buffer

	data := struct {
		IsWorkloadCluster          bool
		SupportsStructuredMetadata bool
	}{
		IsWorkloadCluster: common.IsWorkloadCluster(lc),
		// Promtail in older versions do not support structured metadata.
		SupportsStructuredMetadata: observabilityBundleVersion.GTE(supportsStructuredMetadata),
	}

	err := promtailLoggingConfigTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}
