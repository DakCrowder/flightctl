package authz

import (
	"context"
	"fmt"

	"github.com/flightctl/flightctl/internal/auth/common"
	"github.com/flightctl/flightctl/internal/util"
)

type OIDCAuthZ struct {
	OrgGetter common.OrganizationGetter
}

func NewOIDCAuthZ(orgGetter common.OrganizationGetter) *OIDCAuthZ {
	return &OIDCAuthZ{OrgGetter: orgGetter}
}

func (o OIDCAuthZ) CheckPermission(ctx context.Context, resource string, op string) (bool, error) {
	identity, err := common.GetIdentity(ctx)
	if err != nil {
		return false, err
	}

	orgID, ok := util.GetOrgIdFromContext(ctx)
	if !ok {
		return false, fmt.Errorf("no org id in context")
	}

	org, status := o.OrgGetter.GetOrganization(ctx, orgID)
	if status.Code < 200 || status.Code >= 300 {
		return false, fmt.Errorf("failed to get org %s: %s", orgID, status.Message)
	}

	externalId := org.Spec.ExternalId
	if externalId == nil {
		return false, fmt.Errorf("org %s has no external id", orgID)
	}

	for _, org := range identity.Organizations {
		if org.ID == *externalId {
			return true, nil
		}
	}

	return false, nil
}
