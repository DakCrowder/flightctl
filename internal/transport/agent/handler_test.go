package agenttransport

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"testing"

	"github.com/flightctl/flightctl/internal/consts"
	"github.com/flightctl/flightctl/internal/crypto"
	"github.com/flightctl/flightctl/internal/crypto/signer"
	"github.com/flightctl/flightctl/internal/util"
	"github.com/flightctl/flightctl/pkg/log"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

func createTestCertWithOrgID(orgID uuid.UUID) *x509.Certificate {
	cert := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: "test-device",
		},
		ExtraExtensions: []pkix.Extension{},
	}

	encoded, err := asn1.Marshal(orgID.String())
	if err != nil {
		panic(fmt.Sprintf("failed to marshal org ID: %v", err))
	}

	cert.ExtraExtensions = append(cert.ExtraExtensions, pkix.Extension{
		Id:       signer.OIDOrgID,
		Critical: false,
		Value:    encoded,
	})

	return cert
}

func createTestCertWithInvalidOrgID() *x509.Certificate {
	cert := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: "test-device",
		},
		ExtraExtensions: []pkix.Extension{},
	}

	// Add invalid organization ID extension (not a valid UUID)
	encoded, err := asn1.Marshal("invalid-uuid")
	if err != nil {
		panic(fmt.Sprintf("failed to marshal invalid org ID: %v", err))
	}

	cert.ExtraExtensions = append(cert.ExtraExtensions, pkix.Extension{
		Id:       signer.OIDOrgID,
		Critical: false,
		Value:    encoded,
	})

	return cert
}

func createTestCertWithMalformedExtension() *x509.Certificate {
	cert := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: "test-device",
		},
		ExtraExtensions: []pkix.Extension{},
	}

	// Add malformed extension data
	cert.ExtraExtensions = append(cert.ExtraExtensions, pkix.Extension{
		Id:       signer.OIDOrgID,
		Critical: false,
		Value:    []byte{0x01, 0x02, 0x03}, // Invalid ASN.1 data
	})

	return cert
}

func TestExtractOrgIDFromCertificate(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ExtractOrgIDFromCertificate Suite")
}

var _ = Describe("ExtractOrgIDFromCertificate", func() {
	var (
		ctx     context.Context
		ca      *crypto.CAClient
		logger  logrus.FieldLogger
		testOrg uuid.UUID
	)

	BeforeEach(func() {
		ctx = context.Background()
		ca = &crypto.CAClient{}
		logger = log.NewPrefixLogger("test")
		testOrg = uuid.MustParse("12345678-1234-1234-1234-123456789012")
	})

	Context("when certificate is valid", func() {
		It("should successfully extract organization ID and return context with org ID", func() {
			testCert := createTestCertWithOrgID(testOrg)
			ctxWithCert := context.WithValue(ctx, consts.TLSPeerCertificateCtxKey, testCert)

			resultCtx, err := ExtractOrgIDFromCertificate(ctxWithCert, ca, logger)

			Expect(err).To(BeNil())
			Expect(resultCtx).ToNot(BeNil())

			orgID, ok := util.GetOrgIdFromContext(resultCtx)
			Expect(ok).To(BeTrue())
			Expect(orgID).To(Equal(testOrg))
		})
	})

	Context("when CA client returns an error", func() {
		It("should return wrapped error when PeerCertificateFromCtx fails", func() {
			// Test with context that has no certificate
			resultCtx, err := ExtractOrgIDFromCertificate(ctx, ca, logger)

			Expect(resultCtx).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to get peer certificate from context"))
			Expect(err.Error()).To(ContainSubstring("peer certificate not found"))
		})
	})

	Context("when certificate extensions are invalid", func() {
		It("should return error when organization ID extension is missing", func() {
			testCert := &x509.Certificate{}
			ctxWithCert := context.WithValue(ctx, consts.TLSPeerCertificateCtxKey, testCert)

			resultCtx, err := ExtractOrgIDFromCertificate(ctxWithCert, ca, logger)

			Expect(resultCtx).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to extract organization ID from certificate"))
			Expect(err.Error()).To(ContainSubstring("not found"))
		})

		It("should return error when organization ID has invalid UUID format", func() {
			testCert := createTestCertWithInvalidOrgID()
			ctxWithCert := context.WithValue(ctx, consts.TLSPeerCertificateCtxKey, testCert)

			resultCtx, err := ExtractOrgIDFromCertificate(ctxWithCert, ca, logger)

			Expect(resultCtx).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to parse organization ID from certificate"))
		})

		It("should return error when extension data is malformed", func() {
			testCert := createTestCertWithMalformedExtension()
			ctxWithCert := context.WithValue(ctx, consts.TLSPeerCertificateCtxKey, testCert)

			resultCtx, err := ExtractOrgIDFromCertificate(ctxWithCert, ca, logger)

			Expect(resultCtx).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to extract organization ID from certificate"))
		})
	})

	Context("when certificate is nil", func() {
		It("should return error", func() {
			ctxWithCert := context.WithValue(ctx, consts.TLSPeerCertificateCtxKey, (*x509.Certificate)(nil))

			resultCtx, err := ExtractOrgIDFromCertificate(ctxWithCert, ca, logger)

			Expect(resultCtx).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to get peer certificate from context"))
		})
	})
})
