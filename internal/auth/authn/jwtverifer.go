package authn

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/flightctl/flightctl/internal/auth/common"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type JWTAuth struct {
	oidcAuthority         string
	externalOIDCAuthority string
	jwksUri               string
	clientTlsConfig       *tls.Config
	client                *http.Client
	orgConfig             *common.AuthOrganizationsConfig
}

type OIDCServerResponse struct {
	TokenEndpoint string `json:"token_endpoint"`
	JwksUri       string `json:"jwks_uri"`
}

type JWTUserIdentity struct {
	common.UserIdentity
	organizations map[string]bool
}

func (i *JWTUserIdentity) IsMemberOf(orgID string) bool {
	return i.organizations[orgID]
}

func (i *JWTUserIdentity) SetOrganizations(orgs map[string]bool) {
	i.organizations = orgs
}

func NewJWTAuth(oidcAuthority string, externalOIDCAuthority string, clientTlsConfig *tls.Config, orgConfig *common.AuthOrganizationsConfig) (JWTAuth, error) {
	jwtAuth := JWTAuth{
		oidcAuthority:         oidcAuthority,
		externalOIDCAuthority: externalOIDCAuthority,
		clientTlsConfig:       clientTlsConfig,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: clientTlsConfig,
			},
		},
		orgConfig: orgConfig,
	}

	res, err := jwtAuth.client.Get(fmt.Sprintf("%s/.well-known/openid-configuration", oidcAuthority))
	if err != nil {
		return jwtAuth, err
	}
	oidcResponse := OIDCServerResponse{}
	defer res.Body.Close()
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return jwtAuth, err
	}
	if err := json.Unmarshal(bodyBytes, &oidcResponse); err != nil {
		return jwtAuth, err
	}
	jwtAuth.jwksUri = oidcResponse.JwksUri
	return jwtAuth, nil
}

func (j JWTAuth) ValidateToken(ctx context.Context, token string) error {
	_, err := j.parseAndValidateToken(ctx, token)
	return err
}

func (j JWTAuth) parseAndValidateToken(ctx context.Context, token string) (jwt.Token, error) {
	jwkSet, err := jwk.Fetch(ctx, j.jwksUri, jwk.WithHTTPClient(j.client))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWK set: %w", err)
	}

	parsedToken, err := jwt.Parse([]byte(token), jwt.WithKeySet(jwkSet), jwt.WithValidate(true))
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT token: %w", err)
	}
	return parsedToken, nil
}

func (j JWTAuth) GetIdentity(ctx context.Context, token string) (common.Identity, error) {
	parsedToken, err := j.parseAndValidateToken(ctx, token)
	if err != nil {
		return nil, err
	}

	identity := &JWTUserIdentity{}

	if sub, exists := parsedToken.Get("sub"); exists {
		if uid, ok := sub.(string); ok {
			identity.SetUID(uid)
		}
	}

	if preferredUsername, exists := parsedToken.Get("preferred_username"); exists {
		if username, ok := preferredUsername.(string); ok {
			identity.SetUsername(username)
		}
	}

	orgs := make(map[string]bool)
	if orgClaim, exists := parsedToken.Get("organization"); exists {
		if orgMap, ok := orgClaim.(map[string]interface{}); ok {
			for orgName, orgData := range orgMap {
				orgs[orgName] = true
				if orgDetails, ok := orgData.(map[string]interface{}); ok {
					if id, exists := orgDetails["id"]; exists {
						if idStr, ok := id.(string); ok {
							orgs[idStr] = true
						}
					}
				}
			}
		}
	}
	identity.SetOrganizations(orgs)

	return identity, nil
}

func (j JWTAuth) GetAuthConfig() common.AuthConfig {
	orgConfig := common.AuthOrganizationsConfig{}
	if j.orgConfig != nil {
		orgConfig = *j.orgConfig
	}

	return common.AuthConfig{
		Type:                common.AuthTypeOIDC,
		Url:                 j.externalOIDCAuthority,
		OrganizationsConfig: orgConfig,
	}
}

func (j JWTAuth) GetAuthToken(r *http.Request) (string, error) {
	return common.ExtractBearerToken(r)
}
