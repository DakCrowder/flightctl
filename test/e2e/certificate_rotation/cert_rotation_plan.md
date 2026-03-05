# Agent Certificate Rotation plan

https://issues.redhat.com/browse/EDM-3141


## Goal and background info

Goal: test automatic rotation of the device management certificate performed by the flightctl agent

The device should continue to function before, during, and after certificate rotation without issue or requiring user intervention or re-enrollment

The agent proactively renews the certificate before it expires (lifetime threshold of 75%)

The following environment variables can be used to shorten lifetimes and accelerate rotation for E2E testing:

FLIGHTCTL_TEST_MGMT_CERT_EXPIRY_SECONDS - (API SERVER)
FLIGHTCTL_TEST_CERT_MANAGER_SYNC_INTERVAL - (AGENT)
FLIGHTCTL_TEST_MGMT_CERT_RENEW_BEFORE_SECONDS - (AGENT)
FLIGHTCTL_TEST_MGMT_CERT_RENEW_BEFORE_PERCENT - (AGENT)

## Test outline

Read ./cert_rotation_test_plan.pdf

This test plan should represent a "Describe" Ginkgo block

For each case prefixed with a value like "OCP-87904" that should correspond to an "It" block
Each "It" block should have the numbered value (e.g. 87904 from the above) as a Label like Label("87904")

Populate ./certificate_rotation_suite_test.go with suite logic
- Reference other suite test setups

Populate ./certificate_rotation_test.go with test cases

## Self verification

After writing each test case, you MUST ensure that it passes appropriately.

Running tests:
make run-e2e-test GO_E2E_DIRS=test/e2e/certificate_rotation/
or
make run-e2e-test GINKGO_FOCUS="test name filter goes here"

## Development rules

Rules for development:
- Use Ginkgo framework for outlining tests
- Use Gomega for assertions, all validations must use "Expect" and Gomega matchers

Harness:
Our testing harness in /test/harness/e2e is used to interact with agents on VMs, or connect the server to the local kind k8s cluster
var _ Describe("testing example", func() {
BeforeEach(func() {
    harness := e2e.GetWorkerHarness() // Gets the specific instance 
    login.LoginToAPIWithToken(harness)
})}
When to create a new harness_xxx.go file:
Create a new file when the functions belong to a specific feature (e.g.,harness_rbac.go for RBAC).
If you're adding more than two related Harness methods that don't fit any existing file's scope, create a new harness_xxx.go file.
Note on utility functions:
Helper functions that don't need a *Harness receiver belong in test/util. The harness is for functions that interact with the cluster, devices or VMs.

Test Structure:
All e2e tests must follow this hierarchical structure:
Describe: Top level definition of the feature under test. Define BeforeEach and AfterEach blocks here.
Context: Describes the specific scenario.
It: The specific behavior or assertion being tested. Must reference the Polarion ID.
By: Documentation of steps inside the test. Has no functional impact on test execution

File Organization:
No God Files: Do not add to global utils or harness packages.
Place helper functions that share logic with different files in domain-specific packages (e.g., pkg/e2e/networking/utils.go).
Before you create a function or any line, remember to search the file, someone may have done it before you for easy use. Utils can be found in test/util
All local helper functions must be defined at the very bottom of the file (outside the Describe block).

Label
1. For each Test case, add labels, which are usually "sanity" and the Polarion ID (ask us for it if you don't have access)
2. When adding the "sanity" label, make sure that the overall upstream PR e2e time hasn't increased (<40 minutes). if it is, re-balance/add new test nodes
3. Add the "Agent" label if the test is expected to run for both cs9 and cs10 agents
4. Each "It" should have a Polarion label, if this test is missing in Polarion should be added. Otherwise we use "By"

Structure and readability
1. Constants and variables above, before the test, specs and helper functions - at the bottom.
2. Check for existing constants, especially for things like "time.second*5"
3. Try to consolidate helper functions to avoid repetition.
4. No need for multiple harness calls for each TC- put one harness call in the beforeEach
5. If a string is used more than once - consider putting it in a const.
6. Please, search the harness and testutil for any existing functions, inside of creating new ones or pasting a CLI line.

Code
1. Do not use Expect in functions, only in tests.
2. If a feature is supposed to run on an OCP env, please test it there too.
3. Do not create inline functions inside of tests - write them in the bottom of the file as helper functions.
4. Besides expecting an error not to occur, test for a positive output too [I.E. contain substring("201")]
5. Harden helpers with logging and nil/empty/error cases

Pre-Merge validation:
Code must be tested on OCP env and disconnected env before being merged.
If you don’t have access to an OCP environment:
Contact the QE Team to request a temporary OCP test cluster. 
Coordinate with the Reference QE (the author of the feature tests) assigned to you.
Code must be executed with our current tests to ensure the current tests won’t break

Pull Request (PR) Process:
When you are ready to open your PR, you must follow these steps to get your tests validated and merged:
Add the run-e2e label: Apply this label to your PR. It is required to start the automated E2E test suite.
Request QE review: You must add the Reference QE (or a representative from the QE team) as a reviewer on your PR.
Pass the Suite: The E2E pipeline must complete successfully without any failures before the PR can be merged.

Assertion Standards (Gomega):
We follow an Execute -> Validate Error -> Validate Output pattern. 
Place all assertions (Expect(...)) exclusively within the test blocks (It).(Do not place assertions inside helper functions). 
Pattern: Expecting Success:
// Execute 
harness := e2e.GetWorkerHarness() 
out, err := harness.ManageResource("apply",uniqueFleetYaml)

// Validate Error  
Expect(err).ToNot(HaveOccurred(), "Command failed unexpectedly")

// Validate Output 
Expect(out).To(ContainSubstring("200 OK"), "Output missing expected string")


Pattern: Expecting Failure:
// Execute 
harness := e2e.GetWorkerHarness() 
out, err := harness.ManageResource("apply",badFleetYaml)

// Validate Error 
Expect(err).To(HaveOccurred(), "Command should have failed")

// Validate specific reason in output 
Expect(out).To(ContainSubstring("not found"))


 Reference Implementation:

package e2e_test_example

import (
"github.com/onsi/ginkgo/v2" 
"github.com/onsi/gomega" 
"github.com/flightctl/flightctl/test/harness/e2e" 
"github.com/flightctl/flightctl/test/util"
)

// Constants at the top, before the Describe block
const (
    fleetName    = "test-fleet"
    expectedMsg  = "200 OK"
    errorMsg     = "not found"
)


var _ = Describe("test feature", func() {
    var harness *e2e.Harness

    BeforeEach(func() {
        harness = e2e.GetWorkerHarness()
    })

    Context("Testing example", func() {
      
        It("should show running status", Label( PolarionID ,"sanity"), func() {

            By("Applying a fleet")
     out, err := harness.ManageResource("apply",uniqueFleetYaml)
            Expect(err).ToNot(HaveOccurred(), "Command failed")
            Expect(out).To(ContainSubstring(expectedMsg))
        })

        It("Testing example 2", Label( PolarionID ,"sanity"), func() {

            By("Applying a bad fleetYAML")
     out, err := harness.ManageResource("apply",badFleetYaml)
            Expect(err).To(HaveOccurred(), "Should have failed")
            Expect(out).To(ContainSubstring(errorMsg), "Wrong error msg")
        })
    })
})

// Helper functions, variables, and YAML specs at the bottom
var uniqueFleetYaml = `apiVersion: v1beta1
kind: Fleet
metadata:
  name: test-fleet
`

var badFleetYaml = `invalid: yaml`

