package aap_client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/flightctl/flightctl/internal/auth/common"
	"github.com/flightctl/flightctl/internal/consts"
)

// AAP Gateway API response types
type AAPPaginatedResponse[T any] struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []T     `json:"results"`
}

type AAPTeamSummaryFields struct {
	Organization AAPOrganization `json:"organization"`
}

type AAPTeam struct {
	ID            int                  `json:"id"`
	SummaryFields AAPTeamSummaryFields `json:"summary_fields"`
}

type AAPOrganization struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type AAPOrganizationsResponse = AAPPaginatedResponse[AAPOrganization]
type AAPTeamsResponse = AAPPaginatedResponse[AAPTeam]

type AAPGatewayClient struct {
	gatewayUrl  string
	client      *http.Client
	maxPageSize *int
}

type AAPGatewayClientOptions struct {
	GatewayUrl      string
	ClientTlsConfig *tls.Config
	MaxPageSize     *int
}

func NewAAPGatewayClient(options AAPGatewayClientOptions) *AAPGatewayClient {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: options.ClientTlsConfig,
		},
	}

	return &AAPGatewayClient{
		client:      client,
		gatewayUrl:  options.GatewayUrl,
		maxPageSize: options.MaxPageSize,
	}
}

func (a *AAPGatewayClient) buildURL(path string) string {
	return fmt.Sprintf("%s%s", a.gatewayUrl, path)
}

func getWithPagination[T any](ctx context.Context, a *AAPGatewayClient, path string, token string) ([]T, error) {
	// TODO remove debugging logs
	fmt.Printf("getting with pagination: %s\n", path)

	url := a.buildURL(path)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add(common.AuthHeader, fmt.Sprintf("Bearer %s", token))
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result AAPPaginatedResponse[T]
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	items := result.Results

	if result.Next != nil {
		nextResult, err := getWithPagination[T](ctx, a, *result.Next, token)
		if err != nil {
			return nil, fmt.Errorf("failed to get next page: %w", err)
		}
		items = append(items, nextResult...)
	}

	return items, nil
}

func (a *AAPGatewayClient) appendQueryParams(path string) string {
	if a.maxPageSize != nil {
		return fmt.Sprintf("%s?page_size=%d", path, *a.maxPageSize)
	}
	return path
}

// GET api/gateway/v1/organizations
func (a *AAPGatewayClient) GetOrganizations(ctx context.Context) ([]AAPOrganization, error) {
	path := a.appendQueryParams("/api/gateway/v1/organizations")

	return getWithPagination[AAPOrganization](ctx, a, path, ctx.Value(consts.TokenCtxKey).(string))
}

// GET /api/gateway/v1/users/{user_id}/organizations
func (a *AAPGatewayClient) GetUserOrganizations(ctx context.Context, userID string) ([]AAPOrganization, error) {
	path := a.appendQueryParams(fmt.Sprintf("/api/gateway/v1/users/%s/organizations", userID))

	return getWithPagination[AAPOrganization](ctx, a, path, ctx.Value(consts.TokenCtxKey).(string))
}

// GET /api/gateway/v1/users/{user_id}/teams
func (a *AAPGatewayClient) GetUserTeams(ctx context.Context, userID string) ([]AAPTeam, error) {
	path := a.appendQueryParams(fmt.Sprintf("/api/gateway/v1/users/%s/teams", userID))

	return getWithPagination[AAPTeam](ctx, a, path, ctx.Value(consts.TokenCtxKey).(string))
}
