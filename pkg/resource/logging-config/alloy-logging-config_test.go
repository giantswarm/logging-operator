package loggingconfig

import (
	_ "embed"
	"flag"
	"io/ioutil"
	"path/filepath"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/blang/semver"
	"github.com/google/go-cmp/cmp"

	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	"github.com/giantswarm/logging-operator/pkg/logged-cluster/capicluster"
)

var (
	update = flag.Bool("update", false, "update .golden files")
)

func TestGenerateAlloyLoggingConfig(t *testing.T) {
	testCases := []struct {
		goldenFile                       string
		observabilityBundleVersion       string
		defaultWorkloadClusterNamespaces []string
		installationName                 string
		clusterName                      string
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
			goldenFile:                       "alloy/test/logging-config.alloy.170_MC.yaml",
			observabilityBundleVersion:       "1.7.0",
			defaultWorkloadClusterNamespaces: []string{"test-selector"},
			installationName:                 "test-installation",
			clusterName:                      "test-installation",
		},
		{
			goldenFile:                       "alloy/test/logging-config.alloy.170_WC.yaml",
			observabilityBundleVersion:       "1.7.0",
			defaultWorkloadClusterNamespaces: []string{"test-selector"},
			installationName:                 "test-installation",
			clusterName:                      "test-cluster",
		},
		{
			goldenFile:                       "alloy/test/logging-config.alloy.170_WC_default_namespaces_nil.yaml",
			observabilityBundleVersion:       "1.7.0",
			defaultWorkloadClusterNamespaces: nil,
			installationName:                 "test-installation",
			clusterName:                      "test-cluster",
		},
		{
			goldenFile:                       "alloy/test/logging-config.alloy.170_WC_default_namespaces_empty.yaml",
			observabilityBundleVersion:       "1.7.0",
			defaultWorkloadClusterNamespaces: []string{""},
			installationName:                 "test-installation",
			clusterName:                      "test-cluster",
		},
	}

	for _, tc := range testCases {
		t.Run(filepath.Base(tc.goldenFile), func(t *testing.T) {
			observabilityBundleVersion, err := semver.Parse(tc.observabilityBundleVersion)
			if err != nil {
				t.Fatalf("Failed to parse observability bundle version: %v", err)
			}
			golden, err := ioutil.ReadFile(tc.goldenFile)
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

			config, err := GenerateAlloyLoggingConfig(loggedCluster, observabilityBundleVersion, tc.defaultWorkloadClusterNamespaces)
			if err != nil {
				t.Fatalf("Failed to generate alloy config: %v", err)
			}

			//fmt.Printf("=>> config\n%s", config)
			//fmt.Printf("=>> config_v170_MC\n%s", config_v170_MC)
			if !cmp.Equal(string(golden), config) {
				t.Logf("Generated config differs from %s, diff:\n%s", tc.goldenFile, cmp.Diff(string(golden), config))
				if *update {
					if err := ioutil.WriteFile(tc.goldenFile, []byte(config), 0644); err != nil {
						t.Fatalf("Failed to update golden file: %v", err)
					}
				}
			}
		})
	}
}
