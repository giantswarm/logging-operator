package eventsloggerconfig

import (
	_ "embed"
	"os"
	"path/filepath"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/google/go-cmp/cmp"
)

func TestGenerateAlloyEventsConfig(t *testing.T) {
	testCases := []struct {
		goldenFile        string
		defaultNamespaces []string
		installationName  string
		clusterName       string
		includeNamespaces []string
		excludeNamespaces []string
	}{
		{
			goldenFile:       "alloy/test/events-logger-config.alloy.MC.yaml",
			installationName: "test-installation",
			clusterName:      "test-installation",
		},
		{
			goldenFile:       "alloy/test/events-logger-config.alloy.WC.yaml",
			installationName: "test-installation",
			clusterName:      "test-cluster",
		},
		{
			goldenFile:        "alloy/test/events-logger-config.alloy.WC.include-namespaces.yaml",
			installationName:  "test-installation",
			clusterName:       "include-namespaces",
			includeNamespaces: []string{"namespace1", "namespace2"},
		},
		{
			goldenFile:        "alloy/test/events-logger-config.alloy.WC.exclude-namespaces.yaml",
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

			cluster := &capi.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: tc.clusterName,
				},
			}

			config, err := generateAlloyEventsConfig(cluster, tc.includeNamespaces, tc.excludeNamespaces, tc.installationName, false, "<tempo-url>", []string{"giantswarm"})
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
