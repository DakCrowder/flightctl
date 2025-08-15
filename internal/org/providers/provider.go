package providers

import (
	"context"

	"github.com/flightctl/flightctl/internal/auth/common"
	"github.com/flightctl/flightctl/internal/org"
)

// ExternalOrganizationProvider extracts organization info from various sources
type ExternalOrganizationProvider interface {
	// GetUserOrganizations returns all orgs for a user
	GetUserOrganizations(ctx context.Context, identity common.Identity) ([]org.ExternalOrganization, error)

	// IsMemberOf checks if user is member of specific org
	IsMemberOf(ctx context.Context, identity common.Identity, orgID string) (bool, error)
}
