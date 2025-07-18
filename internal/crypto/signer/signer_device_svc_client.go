package signer

import (
	"context"
	"fmt"
	"strings"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/flightctl/flightctl/internal/util"
	fccrypto "github.com/flightctl/flightctl/pkg/crypto"
	"github.com/google/uuid"
)

const signerDeviceSvcClientExpiryDays int32 = 7

type SignerDeviceSvcClient struct {
	name string
	ca   CA
}

func NewSignerDeviceSvcClient(CAClient CA) Signer {
	cfg := CAClient.Config()
	return &SignerDeviceSvcClient{name: cfg.DeviceSvcClientSignerName, ca: CAClient}
}

func (s *SignerDeviceSvcClient) Name() string {
	return s.name
}

func (s *SignerDeviceSvcClient) Verify(ctx context.Context, request api.CertificateSigningRequest) error {
	cfg := s.ca.Config()

	signer := s.ca.PeerCertificateSignerFromCtx(ctx)

	got := "<nil>"
	if signer != nil {
		got = signer.Name()
	}

	if signer == nil || signer.Name() != cfg.DeviceEnrollmentSignerName {
		return fmt.Errorf("unexpected client certificate signer: expected %q, got %q", cfg.DeviceEnrollmentSignerName, got)
	}

	peerCertificate, err := PeerCertificateFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("failed to get peer certificate from context: %w", err)
	}

	fingerprint, err := DeviceFingerprintFromCN(cfg, peerCertificate.Subject.CommonName)
	if err != nil {
		return fmt.Errorf("failed to extract device fingerprint from peer certificate CN: %w", err)
	}

	parsedCSR, err := fccrypto.ParseCSR(request.Spec.Request)
	if err != nil {
		return fmt.Errorf("failed to parse CSR: %w", err)
	}

	if !strings.HasSuffix(parsedCSR.Subject.CommonName, fmt.Sprintf("-%s", fingerprint)) {
		return fmt.Errorf("CSR CommonName %q does not end with device fingerprint suffix -%s", parsedCSR.Subject.CommonName, fingerprint)
	}

	orgIdFromCtx, ok := util.GetOrgIdFromContext(ctx)
	if !ok {
		return fmt.Errorf("organization ID is required but not found in request context for device enrollment certificate")
	}

	orgIDStr, err := fccrypto.GetExtensionValue(peerCertificate, OIDOrgID)
	if err != nil {
		return fmt.Errorf("organization ID extension not found in peer certificate: %w", err)
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return fmt.Errorf("invalid organization ID format in peer certificate: %w", err)
	}

	if orgID != orgIdFromCtx {
		return fmt.Errorf("organization ID mismatch: peer certificate org ID is %q, context is %q", orgID, orgIdFromCtx)
	}

	return nil
}

func (s *SignerDeviceSvcClient) Sign(ctx context.Context, request api.CertificateSigningRequest) ([]byte, error) {
	cert, err := fccrypto.ParseCSR(request.Spec.Request)
	if err != nil {
		return nil, err
	}

	lastHyphen := strings.LastIndex(cert.Subject.CommonName, "-")
	if lastHyphen == -1 {
		return nil, fmt.Errorf("invalid CN format: no hyphen found in %q", cert.Subject.CommonName)
	}
	fingerprint := cert.Subject.CommonName[lastHyphen+1:]

	orgID, ok := util.GetOrgIdFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("organization ID is required but not found in request context for device enrollment certificate")
	}

	expirySeconds := signerDeviceSvcClientExpiryDays * 24 * 60 * 60
	if request.Spec.ExpirationSeconds != nil && *request.Spec.ExpirationSeconds < expirySeconds {
		expirySeconds = *request.Spec.ExpirationSeconds
	}

	return s.ca.IssueRequestedClientCertificate(
		ctx,
		cert,
		int(expirySeconds),
		WithExtension(OIDOrgID, orgID.String()),
		WithExtension(OIDDeviceFingerprint, fingerprint),
	)
}
