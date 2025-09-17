package common

import (
	"bytes"
	_ "embed"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

var (
	//go:embed templates/alloy-secret.yaml.template
	alloySecretTemplate         string
	compiledAlloySecretTemplate *template.Template
)

func init() {
	compiledAlloySecretTemplate = template.Must(template.New("alloy-secret.yaml").Funcs(sprig.FuncMap()).Parse(alloySecretTemplate))
}

// AlloySecretTemplateData represents the data structure for the Alloy secret template
type AlloySecretTemplateData struct {
	ExtraSecretEnv map[string]string
}

// GenerateAlloySecretValues generates the values for an Alloy secret using the shared template
func GenerateAlloySecretValues(data AlloySecretTemplateData) ([]byte, error) {
	var values bytes.Buffer
	err := compiledAlloySecretTemplate.Execute(&values, data)
	if err != nil {
		return nil, err
	}
	return values.Bytes(), nil
}
