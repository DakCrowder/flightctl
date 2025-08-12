package providers

import "context"

// ExternalOrgProvider fetches external organization IDs that a user has access to.
// Different implementations can fetch from JWTs, external APIs, etc.
type ExternalOrgProvider interface {
	// GetUserOrgs returns external org IDs the user has access to
	GetUserOrgs(ctx context.Context) ([]string, error)

	// HasAccess checks if user has access to a specific external org
	HasAccess(ctx context.Context, externalOrgID string) (bool, error)
}
