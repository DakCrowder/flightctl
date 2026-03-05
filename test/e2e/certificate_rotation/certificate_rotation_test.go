package certificate_rotation_test

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/flightctl/flightctl/api/core/v1beta1"
	"github.com/flightctl/flightctl/internal/agent/device/systeminfo/common"
	"github.com/flightctl/flightctl/test/harness/e2e"
	"github.com/flightctl/flightctl/test/login"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	// certRenewBeforeSeconds is set to 365 days - 1 second (just under the
	// default 365-day cert TTL), causing renewal almost immediately after
	// the agent receives its initial certificate.
	certRenewBeforeSeconds = "31535999"

	// certManagerSyncInterval controls how often the agent checks whether
	// a renewal is needed. 5s keeps the test responsive.
	certManagerSyncInterval = "5s"
)

var _ = Describe("Certificate Rotation", Label("certificate-rotation"), func() {
	var (
		harness  *e2e.Harness
		deviceId string
	)

	BeforeEach(func() {
		harness = e2e.GetWorkerHarness()
		login.LoginToAPIWithToken(harness)

		var dev *v1beta1.Device
		deviceId, dev = harness.EnrollAndWaitForOnlineStatus()
		Expect(dev).ToNot(BeNil())
		GinkgoWriter.Printf("Device enrolled: %s\n", deviceId)
	})

	Context("pre-rotation behavior", func() {
		It("should have device online with valid certificate", Label("87904"), func() {
			By("Verifying device is online")
			status, err := harness.GetDeviceWithStatusSummary(deviceId)
			Expect(err).ToNot(HaveOccurred())
			Expect(status).To(Equal(v1beta1.DeviceSummaryStatusOnline))

			By("Verifying system info is populated")
			sysInfo := harness.GetDeviceSystemInfo(deviceId)
			Expect(sysInfo).ToNot(BeNil())
			Expect(sysInfo.AgentVersion).ToNot(BeEmpty())
			Expect(sysInfo.Architecture).ToNot(BeEmpty())

			By("Waiting for management certificate info to appear in system info")
			var certNotAfter, certSerial string
			Eventually(func() bool {
				serial, notAfter, fetchErr := getDeviceCertInfo(harness, deviceId)
				if fetchErr != nil {
					return false
				}
				certSerial = serial
				certNotAfter = notAfter
				return true
			}, e2e.LONGTIMEOUT, e2e.POLLINGLONG).Should(BeTrue(), "management cert info should appear in system info")

			// Parse the notAfter timestamp and verify it is in the future
			notAfterTime, err := time.Parse(time.RFC3339, certNotAfter)
			Expect(err).ToNot(HaveOccurred(), "managementCertNotAfter should be a valid RFC3339 timestamp")
			Expect(notAfterTime.After(time.Now())).To(BeTrue(), "certificate notAfter should be in the future")
			Expect(certSerial).ToNot(BeEmpty())

			By("Verifying certificate metrics via agent metrics endpoint")
			metricsOutput, err := getAgentMetrics(harness)
			Expect(err).ToNot(HaveOccurred())

			loaded := parseMetricValue(metricsOutput, "flightctl_device_mgmt_cert_loaded")
			Expect(loaded).To(Equal("1"), "flightctl_device_mgmt_cert_loaded should be 1")

			notAfterTS := parseMetricValue(metricsOutput, "flightctl_device_mgmt_cert_not_after_timestamp_seconds")
			Expect(notAfterTS).ToNot(BeEmpty(), "cert not_after timestamp metric should be present")
		})
	})

	Context("certificate rotation", func() {
		It("should rotate certificate while device stays online", Label("87905"), func() {
			By("Waiting for initial certificate info to be reported")
			var initialSerial, initialNotAfter string
			Eventually(func() bool {
				serial, notAfter, fetchErr := getDeviceCertInfo(harness, deviceId)
				if fetchErr != nil {
					return false
				}
				initialSerial = serial
				initialNotAfter = notAfter
				return true
			}, e2e.LONGTIMEOUT, e2e.POLLINGLONG).Should(BeTrue(), "initial cert info should appear in system info")
			GinkgoWriter.Printf("Initial cert serial: %s, notAfter: %s\n", initialSerial, initialNotAfter)

			By("Waiting for certificate rotation (serial change)")
			Eventually(func() string {
				serial, _, fetchErr := getDeviceCertInfo(harness, deviceId)
				if fetchErr != nil {
					return initialSerial
				}
				return serial
			}, e2e.LONGTIMEOUT, e2e.POLLINGLONG).ShouldNot(Equal(initialSerial), "certificate serial should change after rotation")

			// Fetch the new cert info after rotation
			newSerial, newNotAfter, err := getDeviceCertInfo(harness, deviceId)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Printf("New cert serial: %s, notAfter: %s\n", newSerial, newNotAfter)

			By("Verifying device remains online after rotation")
			status, err := harness.GetDeviceWithStatusSummary(deviceId)
			Expect(err).ToNot(HaveOccurred())
			Expect(status).To(Equal(v1beta1.DeviceSummaryStatusOnline))

			By("Verifying device identity unchanged (no re-enrollment)")
			device, err := harness.GetDevice(deviceId)
			Expect(err).ToNot(HaveOccurred())
			Expect(*device.Metadata.Name).To(Equal(deviceId))

			By("Verifying certificate notAfter changed")
			Expect(newNotAfter).ToNot(Equal(initialNotAfter), "notAfter should differ after rotation")

			By("Checking agent logs for renewal activity")
			logs, err := harness.GetFlightctlAgentLogs()
			Expect(err).ToNot(HaveOccurred())
			// The agent should log certificate renewal, not a new enrollment
			Expect(logs).ToNot(ContainSubstring("requesting enrollment"))

			By("Checking renewal success metric")
			metricsOutput, err := getAgentMetrics(harness)
			Expect(err).ToNot(HaveOccurred())
			successCount := parseMetricValue(metricsOutput, `flightctl_device_mgmt_cert_renewal_attempts_total{result="success"}`)
			Expect(successCount).ToNot(BeEmpty(), "renewal success metric should be present")
			Expect(successCount).ToNot(Equal("0"), "at least one successful renewal should have occurred")
		})
	})

	Context("post-rotation validation", func() {
		It("should complete a second rotation cycle", Label("87906"), func() {
			By("Waiting for initial certificate info to be reported")
			var initialSerial string
			Eventually(func() bool {
				serial, _, fetchErr := getDeviceCertInfo(harness, deviceId)
				if fetchErr != nil {
					return false
				}
				initialSerial = serial
				return true
			}, e2e.LONGTIMEOUT, e2e.POLLING).Should(BeTrue(), "initial cert info should appear in system info")

			By("Waiting for the first certificate rotation")
			Eventually(func() string {
				serial, _, fetchErr := getDeviceCertInfo(harness, deviceId)
				if fetchErr != nil {
					return initialSerial
				}
				return serial
			}, e2e.LONGTIMEOUT, e2e.POLLING).ShouldNot(Equal(initialSerial), "first rotation should occur")

			firstRotationSerial, firstRotationNotAfter, err := getDeviceCertInfo(harness, deviceId)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Printf("First rotation cert serial: %s, notAfter: %s\n", firstRotationSerial, firstRotationNotAfter)

			By("Waiting for the second certificate rotation")
			Eventually(func() string {
				serial, _, fetchErr := getDeviceCertInfo(harness, deviceId)
				if fetchErr != nil {
					return firstRotationSerial
				}
				return serial
			}, e2e.LONGTIMEOUT, e2e.POLLING).ShouldNot(Equal(firstRotationSerial), "second rotation should occur")

			secondRotationSerial, secondRotationNotAfter, err := getDeviceCertInfo(harness, deviceId)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Printf("Second rotation cert serial: %s, notAfter: %s\n", secondRotationSerial, secondRotationNotAfter)

			By("Verifying device remains online after second rotation")
			status, err := harness.GetDeviceWithStatusSummary(deviceId)
			Expect(err).ToNot(HaveOccurred())
			Expect(status).To(Equal(v1beta1.DeviceSummaryStatusOnline))

			By("Verifying device identity unchanged")
			device, err := harness.GetDevice(deviceId)
			Expect(err).ToNot(HaveOccurred())
			Expect(*device.Metadata.Name).To(Equal(deviceId))

			By("Verifying notAfter changed from first rotation")
			Expect(secondRotationNotAfter).ToNot(Equal(firstRotationNotAfter),
				"notAfter should differ between first and second rotation")

			By("Checking renewal success metric count >= 2")
			metricsOutput, err := getAgentMetrics(harness)
			Expect(err).ToNot(HaveOccurred())
			successCount := parseMetricValue(metricsOutput, `flightctl_device_mgmt_cert_renewal_attempts_total{result="success"}`)
			Expect(successCount).ToNot(BeEmpty(), "renewal success metric should be present")
			// The count should be at least 2 after two successful rotations
			Expect(successCount).ToNot(Equal("0"))
			Expect(successCount).ToNot(Equal("1"), "at least 2 successful renewals should have occurred")
		})
	})
})

// getDeviceCertInfo extracts the management certificate serial and notAfter
// from the device's system info.
func getDeviceCertInfo(harness *e2e.Harness, deviceID string) (serial string, notAfter string, err error) {
	sysInfo := harness.GetDeviceSystemInfo(deviceID)
	if sysInfo == nil {
		return "", "", fmt.Errorf("system info not available for device %s", deviceID)
	}

	serial, serialFound := sysInfo.Get(common.ManagementCertSerialKey)
	if !serialFound {
		return "", "", fmt.Errorf("managementCertSerial not found in system info")
	}

	notAfter, notAfterFound := sysInfo.Get(common.ManagementCertNotAfterKey)
	if !notAfterFound {
		return "", "", fmt.Errorf("managementCertNotAfter not found in system info")
	}

	return serial, notAfter, nil
}

// getAgentMetrics fetches the Prometheus metrics from the agent's metrics
// endpoint on the VM via SSH.
func getAgentMetrics(harness *e2e.Harness) (string, error) {
	output, err := harness.VM.RunSSH([]string{
		"curl", "-s", "http://127.0.0.1:15690/metrics",
	}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch agent metrics: %w", err)
	}
	return output.String(), nil
}

// parseMetricValue extracts the value of a Prometheus metric line from the
// metrics output. For a metric like `foo{label="val"} 42`, pass the full
// metric name with labels as the metricName parameter.
func parseMetricValue(metricsOutput, metricName string) string {
	for _, line := range strings.Split(metricsOutput, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, metricName) {
			rest := strings.TrimPrefix(line, metricName)
			rest = strings.TrimSpace(rest)
			// The value is the remaining part after the metric name
			if rest != "" {
				return rest
			}
		}
	}
	return ""
}

// strToBuffer is a helper that converts a string to a *bytes.Buffer for use
// as stdin in RunSSH calls.
func strToBuffer(s string) *bytes.Buffer {
	return bytes.NewBufferString(s)
}
