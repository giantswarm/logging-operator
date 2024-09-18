package loggingconfig

import (
	"bytes"
	"text/template"

	_ "embed"

	"github.com/Masterminds/sprig/v3"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

var (
	//go:embed promtail/logging-config.yaml.template
	promtailLoggingConfig         string
	promtailLoggingConfigTemplate *template.Template
)

func init() {
	promtailLoggingConfigTemplate = template.Must(template.New("logging-config.promtail").Funcs(sprig.FuncMap()).Parse(promtailLoggingConfig))
}

// GeneratePromtailLoggingConfig returns a configmap for
// the logging extra-config
func GeneratePromtailLoggingConfig(lc loggedcluster.Interface) (string, error) {
	var values bytes.Buffer

	data := struct {
		IsWorkloadCluster bool
	}{
		IsWorkloadCluster: common.IsWorkloadCluster(lc),
	}

	err := promtailLoggingConfigTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}
