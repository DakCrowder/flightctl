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

type AAPGatewayClient struct {
	gatewayUrl string
	client     *http.Client
}

func NewAAPGatewayClient(gatewayUrl string, clientTlsConfig *tls.Config) *AAPGatewayClient {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: clientTlsConfig,
		},
	}

	return &AAPGatewayClient{
		client:     client,
		gatewayUrl: gatewayUrl,
	}
}

// GET /api/gateway/v1/users/{user_id}/organizations
func (a *AAPGatewayClient) GetUserOrganizations(ctx context.Context, userID string) ([]AAPOrganization, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/gateway/v1/users/%s/organizations", a.gatewayUrl, userID), nil)
	if err != nil {
		return []AAPOrganization{}, fmt.Errorf("failed to create request: %w", err)
	}

	token := ctx.Value(consts.TokenCtxKey)
	if token == nil {
		return []AAPOrganization{}, fmt.Errorf("failed to get token from context")
	}
	req.Header.Add(common.AuthHeader, fmt.Sprintf("Bearer %s", token))

	resp, err := a.client.Do(req)
	if err != nil {
		return []AAPOrganization{}, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return []AAPOrganization{}, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return []AAPOrganization{}, fmt.Errorf("failed to read response body: %w", err)
	}

	organizations := AAPOrganizationsResponse{}
	if err := json.Unmarshal(body, &organizations); err != nil {
		return []AAPOrganization{}, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	// For now ignore pagination and just return results
	return organizations.Results, nil
}
