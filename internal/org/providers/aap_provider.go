package providers

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strconv"

	"github.com/flightctl/flightctl/internal/auth/authn"
	"github.com/flightctl/flightctl/internal/auth/common"
	"github.com/flightctl/flightctl/internal/consts"
	"github.com/flightctl/flightctl/internal/org"
	"github.com/flightctl/flightctl/internal/org/cache"
	"github.com/flightctl/flightctl/pkg/aap_client"
)

type AAPOrganizationProvider struct {
	client *aap_client.AAPGatewayClient
	cache  cache.MembershipCache
}

func cacheKey(userID string, orgID string) string {
	return fmt.Sprintf("%s:%s", userID, orgID)
}

func NewAAPOrganizationProvider(apiUrl string, tlsConfig *tls.Config, cache cache.MembershipCache) (*AAPOrganizationProvider, error) {
	aapClient, err := aap_client.NewAAPGatewayClient(aap_client.AAPGatewayClientOptions{
		GatewayUrl:      apiUrl,
		TLSClientConfig: tlsConfig,
	})
	if err != nil {
		return nil, err
	}

	if cache == nil {
		return nil, fmt.Errorf("AAP organization provider requires a membership cache")
	}

	return &AAPOrganizationProvider{
		client: aapClient,
		cache:  cache,
	}, nil
}

func (p *AAPOrganizationProvider) GetUserOrganizations(ctx context.Context, identity common.Identity) ([]org.ExternalOrganization, error) {
	aapIdentity, ok := identity.(authn.AAPGatewayUserIdentity)
	if !ok {
		return nil, fmt.Errorf("cannot get organizations claims from a non-token identity (got %T)", aapIdentity)
	}

	userID := aapIdentity.GetUID()
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	var orgs []org.ExternalOrganization
	var err error
	if aapIdentity.IsSuperuser() || aapIdentity.IsPlatformAuditor() {
		orgs, err = p.getAllOrganizations(ctx)
	} else {
		orgs, err = p.getUserScopedOrganizations(ctx, userID)
	}

	if err != nil {
		return nil, err
	}
	p.updateCacheFromOrgs(userID, orgs)
	return orgs, nil
}

func (p *AAPOrganizationProvider) getAllOrganizations(ctx context.Context) ([]org.ExternalOrganization, error) {
	token, ok := ctx.Value(consts.TokenCtxKey).(string)
	if !ok {
		return nil, fmt.Errorf("token is required")
	}

	organizations, err := p.client.GetOrganizations(token)
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
	token, ok := ctx.Value(consts.TokenCtxKey).(string)
	if !ok {
		return nil, fmt.Errorf("token is required")
	}

	aapOrganizations, err := p.client.GetUserOrganizations(token, userID)
	if err != nil {
		return nil, err
	}

	aapTeams, err := p.client.GetUserTeams(token, userID)
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

	userID := aapIdentity.GetUID()
	if userID == "" {
		return false, fmt.Errorf("user ID is required")
	}

	if p.cache.Get(cacheKey(userID, externalOrgID)) {
		return true, nil
	}

	var isMember bool
	var err error
	if aapIdentity.IsSuperuser() || aapIdentity.IsPlatformAuditor() {
		isMember, err = p.organizationExists(ctx, externalOrgID)
	} else {
		isMember, err = p.userHasMembership(ctx, userID, externalOrgID)
	}

	if err != nil {
		return false, err
	}

	p.updateCache(userID, externalOrgID, isMember)
	return isMember, nil
}

func (p *AAPOrganizationProvider) organizationExists(ctx context.Context, externalOrgID string) (bool, error) {
	token, ok := ctx.Value(consts.TokenCtxKey).(string)
	if !ok {
		return false, fmt.Errorf("token is required")
	}

	_, err := p.client.GetOrganization(token, externalOrgID)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (p *AAPOrganizationProvider) userHasMembership(ctx context.Context, userID string, externalOrgID string) (bool, error) {
	token, ok := ctx.Value(consts.TokenCtxKey).(string)
	if !ok {
		return false, fmt.Errorf("token is required")
	}

	_, err := p.client.GetOrganization(token, externalOrgID)
	if err != nil && !errors.Is(err, aap_client.ErrNotFound) && !errors.Is(err, aap_client.ErrForbidden) {
		return false, err
	}

	// If we can't get the organization directly, we have to double check team-based membership
	teams, err := p.client.GetUserTeams(token, userID)
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

func (p *AAPOrganizationProvider) updateCache(userID string, externalOrgID string, isMember bool) {
	key := cacheKey(userID, externalOrgID)
	p.cache.Set(key, isMember)
}

func (p *AAPOrganizationProvider) updateCacheFromOrgs(userID string, orgs []org.ExternalOrganization) {
	for _, org := range orgs {
		p.updateCache(userID, org.ID, true)
	}
}
