package authz

import (
	"context"
	"fmt"

	"github.com/flightctl/flightctl/internal/auth/authn"
	"github.com/flightctl/flightctl/internal/auth/common"
	"github.com/flightctl/flightctl/internal/org"
	"github.com/flightctl/flightctl/internal/util"
)

// This const maps to the plural resource name pulled from the url path
const OrganizationsResource = "organizations"

type JWTAuthZ struct {
	orgResolver *org.Resolver
}

func NewJWTAuthZ(orgResolver *org.Resolver) *JWTAuthZ {
	return &JWTAuthZ{
		orgResolver: orgResolver,
	}
}

func (j *JWTAuthZ) CheckPermission(ctx context.Context, resource string, op string) (bool, error) {
	// Skip org validation for listing all organizations - users should be able to
	// list organizations they belong to without requiring a specific org context
	if resource == OrganizationsResource && op == "list" {
		return true, nil
	}

	return j.checkOrgMembership(ctx)
}

func (j *JWTAuthZ) checkOrgMembership(ctx context.Context) (bool, error) {
	identity, err := common.GetIdentity(ctx)
	if err != nil {
		return false, err
	}
	jwtIdentity, ok := identity.(authn.JWTIdentity)
	if !ok {
		return false, fmt.Errorf("identity is not a UserIdentity")
	}

	orgID, ok := util.GetOrgIdFromContext(ctx)
	if !ok {
		return false, fmt.Errorf("org ID not found in context")
	}

	externalID, err := j.orgResolver.GetExternalID(ctx, orgID)
	if err != nil {
		return false, err
	}

	if jwtIdentity.IsMemberOf(string(externalID)) {
		return true, nil
	}

	return false, nil
}
