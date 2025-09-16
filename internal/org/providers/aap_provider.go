package providers

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strconv"

	"github.com/flightctl/flightctl/internal/auth/common"
	"github.com/flightctl/flightctl/internal/org"
	"github.com/flightctl/flightctl/pkg/aap_client"
)

type AAPOrganizationProvider struct {
	client *aap_client.AAPGatewayClient
}

func NewAAPOrganizationProvider(apiUrl string, tlsConfig *tls.Config) (*AAPOrganizationProvider, error) {
	aapClient, err := aap_client.NewAAPGatewayClient(aap_client.AAPGatewayClientOptions{
		GatewayUrl:      apiUrl,
		TLSClientConfig: tlsConfig,
	})
	if err != nil {
		return nil, err
	}

	return &AAPOrganizationProvider{
		client: aapClient,
	}, nil
}

func (p *AAPOrganizationProvider) GetUserOrganizations(ctx context.Context, identity common.Identity) ([]org.ExternalOrganization, error) {
	userID := identity.GetUID()

	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	aapOrganizations, err := p.client.GetUserOrganizations(ctx, userID)
	if err != nil {
		return nil, err
	}

	aapTeams, err := p.client.GetUserTeams(ctx, userID)
	if err != nil {
		return nil, err
	}

	aapOrganizationsMap := make(map[int]*aap_client.AAPOrganization)
	for _, organization := range aapOrganizations {
		aapOrganizationsMap[organization.ID] = organization
	}

	for _, team := range aapTeams {
		aapOrganizationsMap[team.SummaryFields.Organization.ID] = &team.SummaryFields.Organization
	}

	externalOrgs := make([]org.ExternalOrganization, 0, len(aapOrganizationsMap))
	for _, organization := range aapOrganizationsMap {
		externalOrgs = append(externalOrgs, org.ExternalOrganization{
			ID:   strconv.Itoa(organization.ID),
			Name: organization.Name,
		})
	}

	return externalOrgs, nil
}

func (p *AAPOrganizationProvider) IsMemberOf(ctx context.Context, identity common.Identity, externalOrgID string) (bool, error) {
	userID := identity.GetUID()
	if userID == "" {
		return false, fmt.Errorf("user ID is required")
	}

	organization, err := p.client.GetOrganization(ctx, externalOrgID)
	if err != nil {
		if errors.Is(err, aap_client.ErrNotFound) || errors.Is(err, aap_client.ErrForbidden) {
			return false, nil
		}
		return false, err
	}

	if organization == nil {
		return false, nil
	}

	return true, nil
}
