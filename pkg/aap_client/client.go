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
type AAPPaginatedResponse struct {
	Count    int           `json:"count"`
	Next     *string       `json:"next"`
	Previous *string       `json:"previous"`
	Results  []interface{} `json:"results"`
}

type PaginatedResponse[T any] interface {
	GetResults() []T
	GetNext() *string
}

type AAPTeamSummaryFields struct {
	Organization AAPOrganization `json:"organization"`
}

type AAPTeam struct {
	ID            int                  `json:"id"`
	SummaryFields AAPTeamSummaryFields `json:"summary_fields"`
}

func (r AAPTeamsResponse) GetResults() []AAPTeam {
	return r.Results
}

func (r AAPTeamsResponse) GetNext() *string {
	return r.Next
}

type AAPOrganization struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type AAPOrganizationsResponse struct {
	Count    int               `json:"count"`
	Next     *string           `json:"next"`
	Previous *string           `json:"previous"`
	Results  []AAPOrganization `json:"results"`
}

type AAPTeamsResponse struct {
	Count    int       `json:"count"`
	Next     *string   `json:"next"`
	Previous *string   `json:"previous"`
	Results  []AAPTeam `json:"results"`
}

func (r AAPOrganizationsResponse) GetResults() []AAPOrganization {
	return r.Results
}

func (r AAPOrganizationsResponse) GetNext() *string {
	return r.Next
}

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

func getWithPagination[T any, R PaginatedResponse[T]](ctx context.Context, a *AAPGatewayClient, path string, token string, responseFactory func() R) ([]T, error) {
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

	result := responseFactory()
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	items := result.GetResults()

	if result.GetNext() != nil {
		nextResult, err := getWithPagination(ctx, a, *result.GetNext(), token, responseFactory)
		if err != nil {
			return nil, fmt.Errorf("failed to get next page: %w", err)
		}
		items = append(items, nextResult...)
	}

	return items, nil
}

// GET /api/gateway/v1/users/{user_id}/organizations
func (a *AAPGatewayClient) GetUserOrganizations(ctx context.Context, userID string) ([]AAPOrganization, error) {
	var path string
	if a.maxPageSize != nil {
		path = fmt.Sprintf("/api/gateway/v1/users/%s/organizations?page_size=%d", userID, *a.maxPageSize)
	} else {
		path = fmt.Sprintf("/api/gateway/v1/users/%s/organizations", userID)
	}

	return getWithPagination(ctx, a, path, ctx.Value(consts.TokenCtxKey).(string), func() AAPOrganizationsResponse {
		return AAPOrganizationsResponse{}
	})
}

// GET /api/gateway/v1/users/{user_id}/teams
func (a *AAPGatewayClient) GetUserTeams(ctx context.Context, userID string) ([]AAPTeam, error) {
	var path string
	if a.maxPageSize != nil {
		path = fmt.Sprintf("/api/gateway/v1/users/%s/teams?page_size=%d", userID, *a.maxPageSize)
	} else {
		path = fmt.Sprintf("/api/gateway/v1/users/%s/teams", userID)
	}

	return getWithPagination(ctx, a, path, ctx.Value(consts.TokenCtxKey).(string), func() AAPTeamsResponse {
		return AAPTeamsResponse{}
	})
}
