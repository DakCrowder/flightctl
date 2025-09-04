package providers

import (
	"context"

	"github.com/flightctl/flightctl/internal/auth/common"
	"github.com/flightctl/flightctl/internal/org"
)

// Provider types
type AAPGatewayProvider struct {
}

func (p *AAPGatewayProvider) GetUserOrganizations(ctx context.Context, identity common.Identity) ([]org.ExternalOrganization, error) {
	return []org.ExternalOrganization{}, nil
}

func (p *AAPGatewayProvider) IsMemberOf(ctx context.Context, identity common.Identity, externalOrgID string) (bool, error) {
	return false, nil
}
