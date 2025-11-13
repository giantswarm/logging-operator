package eventsloggerconfig

import (
	_ "embed"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/logging-operator/pkg/common"
)

var (
	update = flag.Bool("update", false, "update .golden files")
)

func TestGenerateAlloyEventsConfig(t *testing.T) {
	testCases := []struct {
		goldenFile        string
		defaultNamespaces []string
		installationName  string
		clusterName       string
		includeNamespaces []string
		excludeNamespaces []string
		tracingEnabled    bool
	}{
		{
			goldenFile:       "alloy/test/events-logger-config.alloy.MC.yaml",
			installationName: "test-installation",
			clusterName:      "test-installation",
			tracingEnabled:   false,
		},
		{
			goldenFile:       "alloy/test/events-logger-config.alloy.WC.yaml",
			installationName: "test-installation",
			clusterName:      "test-cluster",
			tracingEnabled:   false,
		},
		{
			goldenFile:        "alloy/test/events-logger-config.alloy.WC.include-namespaces.yaml",
			installationName:  "test-installation",
			clusterName:       "include-namespaces",
			includeNamespaces: []string{"namespace1", "namespace2"},
			tracingEnabled:    false,
		},
		{
			goldenFile:        "alloy/test/events-logger-config.alloy.WC.exclude-namespaces.yaml",
			installationName:  "test-installation",
			clusterName:       "exclude-namespaces",
			excludeNamespaces: []string{"namespace1", "namespace2"},
			tracingEnabled:    false,
		},
		{
			goldenFile:       "alloy/test/events-logger-config.alloy.MC.tracing-enabled.yaml",
			installationName: "test-installation",
			clusterName:      "test-installation",
			tracingEnabled:   true,
		},
		{
			goldenFile:       "alloy/test/events-logger-config.alloy.WC.tracing-enabled.yaml",
			installationName: "test-installation",
			clusterName:      "test-cluster",
			tracingEnabled:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(filepath.Base(tc.goldenFile), func(t *testing.T) {
			golden, err := os.ReadFile(tc.goldenFile)
			if err != nil {
				t.Fatalf("Failed to read golden file: %v", err)
			}

			clusterLabels := common.ClusterLabels{
				ClusterID:    tc.clusterName,
				ClusterType:  "workload_cluster",
				Organization: "test-organization",
				Provider:     "test-provider",
			}
			config, err := generateAlloyEventsConfig(tc.includeNamespaces, tc.excludeNamespaces, false, tc.tracingEnabled, "<tempo-url>", []string{"giantswarm"}, clusterLabels)
			if err != nil {
				t.Fatalf("Failed to generate alloy config: %v", err)
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
