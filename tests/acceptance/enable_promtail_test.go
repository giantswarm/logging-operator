package acceptance_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	gscluster "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/giantswarm/logging-operator/internal/controller"
	loggingreconciler "github.com/giantswarm/logging-operator/pkg/logging-reconciler"
	"github.com/giantswarm/logging-operator/pkg/reconciler"
	promtailtoggle "github.com/giantswarm/logging-operator/pkg/resource/promtail-toggle"
	"github.com/giantswarm/logging-operator/tests"
)

var _ = Describe("Enable Promtail", func() {
	var (
		ctx context.Context

		clusterName    string
		cluster        *gscluster.Cluster
		testReconciler *controller.VintageMCReconciler
		reconcileErr   error
	)

	// Create the dummy cluster with the logging label as well as creates the reconciler
	BeforeEach(func() {
		SetDefaultEventuallyPollingInterval(time.Second)
		SetDefaultEventuallyTimeout(time.Second * 90)

		ctx = context.Background()
		clusterName = tests.GenerateGUID("test")

		// dummy cluster creation
		cluster = &gscluster.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterName,
				Namespace: namespace,
				Labels: map[string]string{
					"giantswarm.io/logging": "true",
				},
			},
			Spec: gscluster.ClusterSpec{
				InfrastructureRef: &corev1.ObjectReference{
					APIVersion: "cluster.x-k8s.io/v1beta1",
					Kind:       "cluster",
					Name:       clusterName,
					Namespace:  namespace,
				},
			},
		}
		Expect(k8sClient.Create(ctx, cluster)).To(Succeed())

		// observability-bundle configmap creation (needed by the reconciler)
		observabilityCm := corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "observability-bundle-user-values",
				Namespace: namespace,
			},
			Data: map[string]string{
				"whatever": "doesn't matter",
			},
		}
		Expect(k8sClient.Create(ctx, &observabilityCm)).To(Succeed())

		// Reconciler creation
		var mgr manager.Manager
		Eventually(func() error {
			var err error
			mgr, err = ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
				Scheme:                 scheme.Scheme,
				MetricsBindAddress:     ":8080",
				Port:                   9443,
				HealthProbeBindAddress: ":8081",
				LeaderElection:         false,
				LeaderElectionID:       "5c8bbafe.x-k8s.io",
			})
			return err
		}).Should(Succeed())

		Expect(mgr).ToNot(BeEmpty())

		promtailReconciler := promtailtoggle.Reconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}

		loggingReconciler := loggingreconciler.LoggingReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
			Reconcilers: []reconciler.Interface{
				&promtailReconciler,
			},
		}

		testReconciler = &controller.VintageMCReconciler{
			Client:            mgr.GetClient(),
			Scheme:            mgr.GetScheme(),
			LoggingReconciler: loggingReconciler,
		}
	})

	// starts the reconciler
	JustBeforeEach(func() {
		request := ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      clusterName,
				Namespace: namespace,
			},
		}
		_, reconcileErr = testReconciler.Reconcile(ctx, request)
	})

	// actual test
	It("enables promtail on the cluster", func() {
		By("creating an  extra-values configmap with promtail enabled in its data")
		configMap := &corev1.ConfigMap{}
		Eventually(func() error {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: "promtail-user-values", Namespace: namespace}, configMap)
			return err
		}).Should(Succeed())

	})

	// cleanup
	When("the cluster is deleted", func() {
		BeforeEach(func() {
			Expect(k8sClient.Delete(ctx, cluster)).To(Succeed())
			Expect(reconcileErr).NotTo(HaveOccurred())
		})

		/* It("removes the finalizers", func() {
			By("")
		}) */
	})
})
