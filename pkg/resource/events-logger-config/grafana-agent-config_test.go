package eventsloggerconfig

import (
	_ "embed"
	"flag"
	"os"
	"path/filepath"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	"github.com/giantswarm/logging-operator/pkg/logged-cluster/capicluster"
)

var (
	update = flag.Bool("update", false, "update .golden files")
)

func TestGenerateGrafanaAgentConfig(t *testing.T) {
	testCases := []struct {
		goldenFile        string
		defaultNamespaces []string
		installationName  string
		clusterName       string
		includeNamespaces []string
		excludeNamespaces []string
	}{
		{
			goldenFile:       "grafana-agent/test/events-logger-config.grafanaagent.MC.yaml",
			installationName: "test-installation",
			clusterName:      "test-installation",
		},
		{
			goldenFile:       "grafana-agent/test/events-logger-config.grafanaagent.WC.yaml",
			installationName: "test-installation",
			clusterName:      "test-cluster",
		},
		{
			goldenFile:        "grafana-agent/test/events-logger-config.grafanaagent.WC.include-namespaces.yaml",
			installationName:  "test-installation",
			clusterName:       "include-namespaces",
			includeNamespaces: []string{"namespace1", "namespace2"},
		},
		{
			goldenFile:        "grafana-agent/test/events-logger-config.grafanaagent.WC.exclude-namespaces.yaml",
			installationName:  "test-installation",
			clusterName:       "exclude-namespaces",
			excludeNamespaces: []string{"namespace1", "namespace2"},
		},
	}

	for _, tc := range testCases {
		t.Run(filepath.Base(tc.goldenFile), func(t *testing.T) {
			golden, err := os.ReadFile(tc.goldenFile)
			if err != nil {
				t.Fatalf("Failed to read golden file: %v", err)
			}

			managementClusterConfig := common.ManagementClusterConfig{
				InstallationName: tc.installationName,
			}

			lc := &capicluster.Object{
				Object: &capi.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: tc.clusterName,
					},
				},
				LoggingAgent: &loggedcluster.LoggingAgent{
					KubeEventsLogger: common.EventsLoggerGrafanaAgent,
				},
			}

			reconciler := &Reconciler{
				ManagementClusterConfig: managementClusterConfig,
			}

			config, err := reconciler.generateGrafanaAgentConfig(lc, tc.includeNamespaces, tc.excludeNamespaces)
			if err != nil {
				t.Fatalf("Failed to generate grafana-agent config: %v", err)
			}

			if string(golden) != config {
				t.Logf("Generated config differs from %s, diff:\n%s", tc.goldenFile, cmp.Diff(string(golden), config))
				t.Fail()
				if *update {
					//nolint:gosec
					if err := os.WriteFile(tc.goldenFile, []byte(config), 0644); err != nil {
						t.Fatalf("Failed to update golden file: %v", err)
					}
				}
			}
		})
	}
}
