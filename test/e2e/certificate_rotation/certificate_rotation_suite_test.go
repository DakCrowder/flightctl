package certificate_rotation_test

import (
	"testing"

	"github.com/flightctl/flightctl/test/harness/e2e"
	testutil "github.com/flightctl/flightctl/test/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCertificateRotation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Certificate Rotation E2E Suite")
}

var _ = BeforeSuite(func() {
	_, _, err := e2e.SetupWorkerHarness()
	Expect(err).ToNot(HaveOccurred())
})

var _ = BeforeEach(func() {
	workerID := GinkgoParallelProcess()
	harness := e2e.GetWorkerHarness()
	suiteCtx := e2e.GetWorkerContext()

	GinkgoWriter.Printf("BeforeEach Worker %d: Setting up test with VM from pool (cert rotation)\n", workerID)

	ctx := testutil.StartSpecTracerForGinkgo(suiteCtx)
	harness.SetTestContext(ctx)

	// Use SetupVMFromPool (not SetupVMFromPoolAndStartAgent) so we can
	// inject environment variables via a systemd drop-in before the agent starts.
	err := harness.SetupVMFromPool(workerID)
	Expect(err).ToNot(HaveOccurred())

	// Create systemd drop-in to configure accelerated certificate rotation
	err = createCertRotationDropIn(harness)
	Expect(err).ToNot(HaveOccurred())

	// Enable the agent metrics endpoint for test validation
	err = enableAgentMetrics(harness)
	Expect(err).ToNot(HaveOccurred())

	// Reload systemd and start the agent with the test configuration
	_, err = harness.VM.RunSSH([]string{"sudo", "systemctl", "daemon-reload"}, nil)
	Expect(err).ToNot(HaveOccurred())

	err = harness.StartFlightCtlAgent()
	Expect(err).ToNot(HaveOccurred())

	GinkgoWriter.Printf("BeforeEach Worker %d: Test setup completed (cert rotation)\n", workerID)
})

var _ = AfterEach(func() {
	workerID := GinkgoParallelProcess()
	GinkgoWriter.Printf("AfterEach Worker %d: Cleaning up test resources\n", workerID)

	harness := e2e.GetWorkerHarness()
	suiteCtx := e2e.GetWorkerContext()

	harness.PrintAgentLogsIfFailed()

	err := harness.CleanUpAllTestResources()
	Expect(err).ToNot(HaveOccurred())

	harness.SetTestContext(suiteCtx)

	GinkgoWriter.Printf("AfterEach Worker %d: Test cleanup completed\n", workerID)
})

// createCertRotationDropIn creates a systemd drop-in that sets the environment
// variables needed to accelerate certificate renewal for testing.
func createCertRotationDropIn(harness *e2e.Harness) error {
	dropInContent := `[Service]
Environment="FLIGHTCTL_TEST_CERT_MANAGER_SYNC_INTERVAL=` + certManagerSyncInterval + `"
Environment="FLIGHTCTL_TEST_MGMT_CERT_RENEW_BEFORE_SECONDS=` + certRenewBeforeSeconds + `"
`
	// Create the drop-in directory
	if _, err := harness.VM.RunSSH([]string{
		"sudo", "mkdir", "-p", "/etc/systemd/system/flightctl-agent.service.d",
	}, nil); err != nil {
		return err
	}

	// Write the drop-in file via sudo tee
	if _, err := harness.VM.RunSSH([]string{
		"sudo", "tee", "/etc/systemd/system/flightctl-agent.service.d/cert-rotation-test.conf",
	}, strToBuffer(dropInContent)); err != nil {
		return err
	}

	return nil
}

// enableAgentMetrics appends metrics-enabled: true to the agent config so
// the Prometheus /metrics endpoint is available for test validation.
func enableAgentMetrics(harness *e2e.Harness) error {
	metricsConfig := "\nmetrics-enabled: true\n"
	_, err := harness.VM.RunSSH([]string{
		"sudo", "tee", "-a", "/etc/flightctl/config.yaml",
	}, strToBuffer(metricsConfig))
	return err
}
