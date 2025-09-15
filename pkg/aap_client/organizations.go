package aap_client

import (
	"context"
	"fmt"

	"github.com/flightctl/flightctl/internal/consts"
)

type AAPOrganization struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type AAPOrganizationsResponse = AAPPaginatedResponse[AAPOrganization]

// GET /api/gateway/v1/organizations/{organization_id}
func (a *AAPGatewayClient) GetOrganization(ctx context.Context, organizationID string) (*AAPOrganization, error) {
	path := a.appendQueryParams(fmt.Sprintf("/api/gateway/v1/organizations/%s", organizationID))
	// TODO shold we return pointers or values consistently?
	// TODO handle error if token is not in context
	result, err := get[AAPOrganization](a, path, ctx.Value(consts.TokenCtxKey).(string))
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GET /api/gateway/v1/organizations
func (a *AAPGatewayClient) GetOrganizations(ctx context.Context) ([]*AAPOrganization, error) {
	path := a.appendQueryParams("/api/gateway/v1/organizations")
	return getWithPagination[AAPOrganization](a, path, ctx.Value(consts.TokenCtxKey).(string))
}

// GET /api/gateway/v1/users/{user_id}/organizations
func (a *AAPGatewayClient) GetUserOrganizations(ctx context.Context, userID string) ([]*AAPOrganization, error) {
	path := a.appendQueryParams(fmt.Sprintf("/api/gateway/v1/users/%s/organizations", userID))
	return getWithPagination[AAPOrganization](a, path, ctx.Value(consts.TokenCtxKey).(string))
}
