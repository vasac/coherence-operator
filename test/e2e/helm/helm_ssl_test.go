package helm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/oracle/coherence-operator/test/e2e/helper/matchers"

	"github.com/oracle/coherence-operator/test/e2e/helper"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"time"
)

var _ = Describe("Operator Helm Chart with SSL", func() {
	var (
		hm *helper.HelmReleaseManager

		secretName = "test-ssl-secret"
		keyFile    = "test.key"
		certFile   = "test.cert"
		caFile     = "test.ca"
	)

	When("installing Helm chart with SSL values", func() {

		// The JustBefore function is where the Helm install happens
		JustBeforeEach(func() {
			// create a dummy secret
			err := helper.EnsureSecretDeleted(HelmHelper.KubeClient, HelmHelper.Namespace, secretName)
			ssl, _, err := helper.CreateSslSecret(HelmHelper.KubeClient, HelmHelper.Namespace, secretName)
			Expect(err).ToNot(HaveOccurred())

			// Create the values to use
			values := helper.OperatorValues{CoherenceOperator: &helper.OperatorSpec{SSL: ssl}}

			// Create a HelmReleaseManager with a release name and values
			hm, err = HelmHelper.NewOperatorHelmReleaseManager("ssl-test", &values)
			Expect(err).ToNot(HaveOccurred())

			// Install the chart
			_, err = hm.InstallRelease()
			Expect(err).ToNot(HaveOccurred())
		})

		// The JustAfter function will ensure the chart is uninstalled
		JustAfterEach(func() {
			// ensure that the chart is uninstalled
			_, err := hm.UninstallRelease()
			Expect(err).ToNot(HaveOccurred())

			err = helper.WaitForOperatorCleanup(HelmHelper.KubeClient, HelmHelper.Namespace, GinkgoT())
			Expect(err).ToNot(HaveOccurred())

			// delete the ssl secret
			err = HelmHelper.KubeClient.CoreV1().Secrets(HelmHelper.Namespace).Delete(secretName, &metav1.DeleteOptions{})
			Expect(err).ToNot(HaveOccurred())
		})

		It("should deploy the Operator with SSL environment variables set", func() {
			// The chart is installed but the Pod(s) may not exist yet so wait for it...
			// (we wait a maximum of 5 minutes, retrying every 10 seconds)
			pods, err := helper.WaitForOperatorPods(HelmHelper.KubeClient, HelmHelper.Namespace, time.Second*10, time.Minute*5)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(pods)).To(Equal(1))

			// Assert SSL environment variables
			container := pods[0].Spec.Containers[0]
			Expect(container.Env).To(HaveEnvVar(corev1.EnvVar{Name: "SSL_CERTS_DIR", Value: "/coherence/certs"}))
			Expect(container.Env).To(HaveEnvVar(corev1.EnvVar{Name: "SSL_KEY_FILE", Value: keyFile}))
			Expect(container.Env).To(HaveEnvVar(corev1.EnvVar{Name: "SSL_CERT_FILE", Value: certFile}))
			Expect(container.Env).To(HaveEnvVar(corev1.EnvVar{Name: "SSL_CA_FILE", Value: caFile}))
		})
	})
})