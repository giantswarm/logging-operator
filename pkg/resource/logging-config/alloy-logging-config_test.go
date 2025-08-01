package loggingconfig

import (
	_ "embed"
	"flag"
	"os"
	"path/filepath"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/blang/semver"
	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/logging-operator/pkg/common"
)

var (
	update = flag.Bool("update", false, "update .golden files")
)

func TestGenerateAlloyLoggingConfig(t *testing.T) {
	testCases := []struct {
		goldenFile                 string
		observabilityBundleVersion string
		defaultNamespaces          []string
		installationName           string
		clusterName                string
		tenants                    []string
	}{
		{
			goldenFile:                 "alloy/test/logging-config.alloy.162_MC.yaml",
			observabilityBundleVersion: "1.6.2",
			installationName:           "test-installation",
			clusterName:                "test-installation",
		},
		{
			goldenFile:                 "alloy/test/logging-config.alloy.162_WC.yaml",
			observabilityBundleVersion: "1.6.2",
			installationName:           "test-installation",
			clusterName:                "test-cluster",
		},
		{
			goldenFile:                 "alloy/test/logging-config.alloy.170_MC.yaml",
			observabilityBundleVersion: "1.7.0",
			defaultNamespaces:          []string{"test-selector"},
			installationName:           "test-installation",
			clusterName:                "test-installation",
		},
		{
			goldenFile:                 "alloy/test/logging-config.alloy.170_WC.yaml",
			observabilityBundleVersion: "1.7.0",
			defaultNamespaces:          []string{"test-selector"},
			installationName:           "test-installation",
			clusterName:                "test-cluster",
		},
		{
			goldenFile:                 "alloy/test/logging-config.alloy.170_WC_default_namespaces_nil.yaml",
			observabilityBundleVersion: "1.7.0",
			defaultNamespaces:          nil,
			installationName:           "test-installation",
			clusterName:                "test-cluster",
		},
		{
			goldenFile:                 "alloy/test/logging-config.alloy.170_WC_default_namespaces_empty.yaml",
			observabilityBundleVersion: "1.7.0",
			defaultNamespaces:          []string{""},
			installationName:           "test-installation",
			clusterName:                "test-cluster",
		},
		{
			goldenFile:                 "alloy/test/logging-config.alloy.170_WC_custom_tenants.yaml",
			observabilityBundleVersion: "1.7.0",
			defaultNamespaces:          []string{""},
			installationName:           "test-installation",
			clusterName:                "test-cluster",
			tenants:                    []string{"test-tenant-a", "test-tenant-b"},
		},
	}

	for _, tc := range testCases {
		t.Run(filepath.Base(tc.goldenFile), func(t *testing.T) {
			observabilityBundleVersion, err := semver.Parse(tc.observabilityBundleVersion)
			if err != nil {
				t.Fatalf("Failed to parse observability bundle version: %v", err)
			}
			golden, err := os.ReadFile(tc.goldenFile)
			if err != nil {
				t.Fatalf("Failed to read golden file: %v", err)
			}

			cluster := &capi.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: tc.clusterName,
				},
			}

			loggingAgent := &common.LoggingAgent{
				LoggingAgent:     common.LoggingAgentAlloy,
				KubeEventsLogger: common.EventsLoggerAlloy,
			}

			config, err := GenerateAlloyLoggingConfig(cluster, loggingAgent, observabilityBundleVersion, tc.defaultNamespaces, tc.tenants, tc.installationName, false)
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
