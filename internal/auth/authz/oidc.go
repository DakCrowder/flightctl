package authz

import (
	"context"
	"fmt"

	"github.com/flightctl/flightctl/internal/auth/common"
	"github.com/flightctl/flightctl/internal/org"
	"github.com/flightctl/flightctl/internal/util"
)

const (
	// This const maps to the plural resource name pulled from the url path
	OrganizationsResource = "organizations"
)

type OIDCAuthZ struct {
	orgResolver *org.Resolver
}

func NewOIDCAuthZ(orgResolver *org.Resolver) *OIDCAuthZ {
	return &OIDCAuthZ{
		orgResolver: orgResolver,
	}
}

func (o OIDCAuthZ) CheckPermission(ctx context.Context, resource string, op string) (bool, error) {
	// Skip org validation for listing all organizations - users should be able to
	// list organizations they belong to without requiring a specific org context
	if resource == OrganizationsResource && op == "list" {
		return true, nil
	}

	return o.CheckOrgPermission(ctx)
}

func (o OIDCAuthZ) CheckOrgPermission(ctx context.Context) (bool, error) {
	identity, err := common.GetIdentity(ctx)
	if err != nil {
		return false, err
	}

	orgID, ok := util.GetOrgIdFromContext(ctx)
	if !ok {
		return false, fmt.Errorf("no org id in context")
	}

	externalID, err := o.orgResolver.GetExternalID(ctx, orgID)
	if err != nil {
		return false, err
	}

	for _, org := range identity.Organizations {
		if org.ID == externalID {
			return true, nil
		}
	}

	return false, nil
}
