package providers

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strconv"

	"github.com/flightctl/flightctl/internal/auth/authn"
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
	aapIdentity, ok := identity.(authn.AAPGatewayUserIdentity)
	if !ok {
		return nil, fmt.Errorf("cannot get organizations claims from a non-token identity (got %T)", aapIdentity)
	}

	if aapIdentity.IsSuperuser() || aapIdentity.IsPlatformAuditor() {
		return p.getAllOrganizations(ctx)
	}

	userID := aapIdentity.GetUID()

	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	return p.getUserScopedOrganizations(ctx, userID)
}

func (p *AAPOrganizationProvider) getAllOrganizations(ctx context.Context) ([]org.ExternalOrganization, error) {
	organizations, err := p.client.GetOrganizations(ctx)
	if err != nil {
		return nil, err
	}
	externalOrgs := make([]org.ExternalOrganization, 0, len(organizations))
	for _, organization := range organizations {
		externalOrgs = append(externalOrgs, org.ExternalOrganization{
			ID:   strconv.Itoa(organization.ID),
			Name: organization.Name,
		})
	}

	return externalOrgs, nil
}

func (p *AAPOrganizationProvider) getUserScopedOrganizations(ctx context.Context, userID string) ([]org.ExternalOrganization, error) {
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
	aapIdentity, ok := identity.(authn.AAPGatewayUserIdentity)
	if !ok {
		return false, fmt.Errorf("cannot get organizations claims from a non-token identity (got %T)", aapIdentity)
	}

	if aapIdentity.IsSuperuser() || aapIdentity.IsPlatformAuditor() {
		return p.organizationExists(ctx, externalOrgID)
	}

	userID := aapIdentity.GetUID()
	if userID == "" {
		return false, fmt.Errorf("user ID is required")
	}

	return p.userHasMembership(ctx, userID, externalOrgID)
}

func (p *AAPOrganizationProvider) organizationExists(ctx context.Context, externalOrgID string) (bool, error) {
	_, err := p.client.GetOrganization(ctx, externalOrgID)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (p *AAPOrganizationProvider) userHasMembership(ctx context.Context, userID string, externalOrgID string) (bool, error) {
	_, err := p.client.GetOrganization(ctx, externalOrgID)
	if err != nil && !errors.Is(err, aap_client.ErrNotFound) && !errors.Is(err, aap_client.ErrForbidden) {
		return false, err
	}

	// If we can't get the organization directly, we have to double check team-based membership
	teams, err := p.client.GetUserTeams(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, team := range teams {
		if strconv.Itoa(team.SummaryFields.Organization.ID) == externalOrgID {
			return true, nil
		}
	}

	return false, nil
}
