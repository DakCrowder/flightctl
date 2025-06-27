package authz

import (
	"context"
	"errors"

	"github.com/flightctl/flightctl/internal/flterrors"
	"github.com/flightctl/flightctl/internal/store"
	"github.com/flightctl/flightctl/internal/util"
)

// BasicOrgAuth is an AuthZ provider that validates organization access
// It checks if the organization in the request context exists in the database
type BasicOrgAuth struct {
	store store.Store
}

func NewBasicOrgAuth(store store.Store) *BasicOrgAuth {
	return &BasicOrgAuth{store: store}
}

func (a *BasicOrgAuth) CheckPermission(ctx context.Context, resource string, op string) (bool, error) {
	orgID, ok := util.GetOrgIdFromContext(ctx)
	if !ok {
		return false, flterrors.ErrInvalidOrganizationID
	}

	if a.store == nil {
		return false, nil
	}

	_, err := a.store.Organization().Get(ctx, orgID)
	if err != nil {
		if errors.Is(err, flterrors.ErrResourceNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
