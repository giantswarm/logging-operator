package loggingconfig

import (
	_ "embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/blang/semver"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	"github.com/giantswarm/logging-operator/pkg/logged-cluster/capicluster"
)

func TestGeneratePromtailLoggingConfig(t *testing.T) {
	testCases := []struct {
		goldenFile                 string
		observabilityBundleVersion string
		installationName           string
		clusterName                string
	}{
		{
			goldenFile:                 "promtail/test/logging-config.promtail.090_MC.yaml",
			observabilityBundleVersion: "0.9.0",
			installationName:           "test-installation",
			clusterName:                "test-installation",
		},
		{
			goldenFile:                 "promtail/test/logging-config.promtail.090_WC.yaml",
			observabilityBundleVersion: "0.9.0",
			installationName:           "test-installation",
			clusterName:                "test-cluster",
		},
		{
			goldenFile:                 "promtail/test/logging-config.promtail.100_MC.yaml",
			observabilityBundleVersion: "1.0.0",
			installationName:           "test-installation",
			clusterName:                "test-installation",
		},
		{
			goldenFile:                 "promtail/test/logging-config.promtail.100_WC.yaml",
			observabilityBundleVersion: "1.0.0",
			installationName:           "test-installation",
			clusterName:                "test-cluster",
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

			loggedCluster := &capicluster.Object{
				Object: &capi.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: tc.clusterName,
					},
				},
				Options: loggedcluster.Options{
					InstallationName: tc.installationName,
				},
			}

			config, err := GeneratePromtailLoggingConfig(loggedCluster, observabilityBundleVersion)
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