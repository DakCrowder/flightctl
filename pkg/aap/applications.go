package aap

import (
	"context"
	"fmt"
	"net/url"
)

type AAPOAuthApplicationRequest struct {
	Name                   string `json:"name"`
	Organization           int    `json:"organization"`
	AuthorizationGrantType string `json:"authorization_grant_type"`
	ClientType             string `json:"client_type"`
	RedirectURIs           string `json:"redirect_uris"`
	AppURL                 string `json:"app_url"`
}

type AAPOAuthApplicationResponse struct {
	ID                     int    `json:"id"`
	Name                   string `json:"name"`
	ClientID               string `json:"client_id"`
	ClientSecret           string `json:"client_secret,omitempty"`
	ClientType             string `json:"client_type"`
	AuthorizationGrantType string `json:"authorization_grant_type"`
	RedirectURIs           string `json:"redirect_uris"`
	AppURL                 string `json:"app_url"`
	Organization           int    `json:"organization"`
}

// CreateOAuthApplication creates a new OAuth application in AAP Gateway
// POST /api/gateway/v1/applications/
func (a *AAPGatewayClient) CreateOAuthApplication(ctx context.Context, token string, request *AAPOAuthApplicationRequest) (*AAPOAuthApplicationResponse, error) {
	endpoint := a.buildEndpoint("/api/gateway/v1/applications/", nil)
	return post[AAPOAuthApplicationResponse](a, ctx, endpoint, token, request)
}

// GetOAuthApplicationByName looks up an OAuth application by name and organization.
// GET /api/gateway/v1/applications/?name=<name>&organization=<org>
func (a *AAPGatewayClient) GetOAuthApplicationByName(ctx context.Context, token string, name string, organization int) (*AAPOAuthApplicationResponse, error) {
	query := url.Values{}
	query.Set("name", name)
	query.Set("organization", fmt.Sprintf("%d", organization))
	endpoint := a.buildEndpoint("/api/gateway/v1/applications/", query)
	results, err := getWithPagination[AAPOAuthApplicationResponse](a, ctx, endpoint, token)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, ErrNotFound
	}
	return results[0], nil
}
