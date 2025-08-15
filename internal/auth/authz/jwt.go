package authz

import (
	"context"
	"fmt"

	"github.com/flightctl/flightctl/internal/auth/common"
	"github.com/flightctl/flightctl/internal/util"
	"github.com/google/uuid"
)

// This const maps to the plural resource name pulled from the url path
const OrganizationsResource = "organizations"

type MembershipChecker interface {
	IsMemberOf(ctx context.Context, identity common.Identity, orgID uuid.UUID) (bool, error)
}

type JWTAuthZ struct {
	membershipChecker MembershipChecker
}

func NewJWTAuthZ(checker MembershipChecker) *JWTAuthZ {
	return &JWTAuthZ{
		membershipChecker: checker,
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
		return false, fmt.Errorf("failed to get identity: %w", err)
	}

	orgID, ok := util.GetOrgIdFromContext(ctx)
	if !ok {
		return false, fmt.Errorf("org ID not found in context")
	}

	isMember, err := j.membershipChecker.IsMemberOf(ctx, identity, orgID)
	if err != nil {
		return false, fmt.Errorf("failed to check org membership: %w", err)
	}
	return isMember, nil
}
