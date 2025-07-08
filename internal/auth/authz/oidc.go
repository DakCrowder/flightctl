package authz

import (
	"context"
	"errors"

	"github.com/flightctl/flightctl/internal/auth/common"
	"github.com/flightctl/flightctl/internal/flterrors"
	"github.com/flightctl/flightctl/internal/util"
	"github.com/samber/lo"
)

type OIDCAuthz struct {
	orgGetter OrganizationGetter
}

func NewOIDCAuthz(orgGetter OrganizationGetter) *OIDCAuthz {
	return &OIDCAuthz{orgGetter: orgGetter}
}

func (a *OIDCAuthz) CheckPermission(ctx context.Context, resource string, op string) (bool, error) {
	orgID, ok := util.GetOrgIdFromContext(ctx)
	if !ok {
		return false, flterrors.ErrInvalidOrganizationID
	}

	if a.orgGetter == nil {
		return false, nil
	}

	organization, err := a.orgGetter.Get(ctx, orgID)
	if err != nil {
		if errors.Is(err, flterrors.ErrResourceNotFound) {
			return false, nil
		}
		return false, err
	}

	identity, err := common.GetIdentity(ctx)
	if err != nil {
		return false, err
	}

	if !identity.Organizations[lo.FromPtr(organization.ExternalId)] {
		return false, nil
	}

	return true, nil
}
