package authz

import (
	"context"
	"errors"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/flightctl/flightctl/internal/flterrors"
	"github.com/flightctl/flightctl/internal/util"
	"github.com/google/uuid"
)

// OrganizationGetter defines the interface for retrieving organizations
type OrganizationGetter interface {
	Get(ctx context.Context, orgID uuid.UUID) (*api.Organization, error)
}

// BasicOrgAuth is an AuthZ provider that validates organization access
// It checks if the organization in the request context exists in the database
type BasicOrgAuth struct {
	orgGetter OrganizationGetter
}

func NewBasicOrgAuth(orgGetter OrganizationGetter) *BasicOrgAuth {
	return &BasicOrgAuth{orgGetter: orgGetter}
}

func (a *BasicOrgAuth) CheckPermission(ctx context.Context, resource string, op string) (bool, error) {
	orgID, ok := util.GetOrgIdFromContext(ctx)
	if !ok {
		return false, flterrors.ErrInvalidOrganizationID
	}

	if a.orgGetter == nil {
		return false, nil
	}

	_, err := a.orgGetter.Get(ctx, orgID)
	if err != nil {
		if errors.Is(err, flterrors.ErrResourceNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
