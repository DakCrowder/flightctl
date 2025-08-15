package providers

import (
	"context"
	"fmt"

	"github.com/flightctl/flightctl/internal/auth/authn"
	"github.com/flightctl/flightctl/internal/auth/common"
	"github.com/flightctl/flightctl/internal/org"
)

const OrganizationClaimName = "organization"

type ClaimsProvider struct{}

func (c *ClaimsProvider) GetUserOrganizations(ctx context.Context, identity common.Identity) ([]org.ExternalOrganization, error) {
	orgs, err := claimsFromIdentity(identity)
	if err != nil {
		return nil, err
	}

	externalOrgs := make([]org.ExternalOrganization, 0, len(orgs))
	for orgName, orgID := range orgs {
		externalOrgs = append(externalOrgs, org.ExternalOrganization{
			ID:   orgID,
			Name: orgName,
		})
	}

	return externalOrgs, nil
}

func (c *ClaimsProvider) IsMemberOf(ctx context.Context, identity common.Identity, orgID string) (bool, error) {
	orgs, err := claimsFromIdentity(identity)
	if err != nil {
		return false, err
	}

	for _, claimID := range orgs {
		if claimID == orgID {
			return true, nil
		}
	}

	return false, nil
}

func claimsFromIdentity(identity common.Identity) (map[string]string, error) {
	tokenIdentity, ok := identity.(authn.TokenIdentity)
	if !ok {
		return nil, fmt.Errorf("cannot get organizations claims from a non-token identity")
	}

	organizationClaims, ok := tokenIdentity.GetClaim(OrganizationClaimName)
	if !ok {
		return nil, fmt.Errorf("unable to get organizations claims from token identity")
	}

	// Organization claims are a map of orgName -> orgID
	orgs, ok := organizationClaims.(map[string]string)
	if !ok {
		return nil, fmt.Errorf("invalid organizations claims format")
	}

	return orgs, nil
}
