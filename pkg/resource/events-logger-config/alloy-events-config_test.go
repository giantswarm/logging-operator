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

	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	"github.com/giantswarm/logging-operator/pkg/logged-cluster/capicluster"
)

var (
	isUpdate = flag.Bool("update", false, "update .golden files")
)

func TestGenerateAlloyEventsConfig(t *testing.T) {
	testCases := []struct {
		goldenFile        string
		defaultNamespaces []string
		installationName  string
		clusterName       string
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
	}

	for _, tc := range testCases {
		t.Run(filepath.Base(tc.goldenFile), func(t *testing.T) {
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
					KubeEventsLogger: "alloy",
				},
			}

			config, err := generateAlloyEventsConfig(loggedCluster, []string{"kube-system", "giantswarm"})
			if err != nil {
				t.Fatalf("Failed to generate alloy config: %v", err)
			}

			if string(golden) != config {
				t.Logf("Generated config differs from %s, diff:\n%s", tc.goldenFile, cmp.Diff(string(golden), config))
				t.Fail()
				if *isUpdate {
					//nolint:gosec
					if err := os.WriteFile(tc.goldenFile, []byte(config), 0644); err != nil {
						t.Fatalf("Failed to update golden file: %v", err)
					}
				}
			}
		})
	}
}
