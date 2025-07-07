package loggingconfig

import (
	"bytes"
	_ "embed"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/logging-operator/pkg/common"
)

var (
	//go:embed promtail/logging-config.promtail.yaml.template
	promtailLoggingConfig         string
	promtailLoggingConfigTemplate *template.Template
)

func init() {
	promtailLoggingConfigTemplate = template.Must(template.New("logging-config.promtail.yaml").Funcs(sprig.FuncMap()).Parse(promtailLoggingConfig))
}

// GeneratePromtailLoggingConfig returns a configmap for
// the logging extra-config
func GeneratePromtailLoggingConfig(cluster *capi.Cluster, installationName string) (string, error) {
	var values bytes.Buffer

	data := struct {
		IsWorkloadCluster bool
	}{
		IsWorkloadCluster: common.IsWorkloadCluster(installationName, cluster.GetName()),
	}

	err := promtailLoggingConfigTemplate.Execute(&values, data)
	if err != nil {
		return "", err
	}

	return values.String(), nil
}
