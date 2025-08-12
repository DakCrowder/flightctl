package common

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/flightctl/flightctl/internal/org"
	"github.com/flightctl/flightctl/internal/util"
	"github.com/google/uuid"
)

type ctxKeyAuthHeader string

const (
	AuthHeader     string           = "Authorization"
	TokenCtxKey    ctxKeyAuthHeader = "TokenCtxKey"
	IdentityCtxKey ctxKeyAuthHeader = "IdentityCtxKey"
)

const (
	AuthTypeK8s  = "k8s"
	AuthTypeOIDC = "OIDC"
	AuthTypeAAP  = "AAPGateway"
)

type AuthConfig struct {
	Type string
	Url  string
}

type Identity struct {
	Username      string
	UID           string
	Groups        []string
	Organizations []ExternalOrganization
}

type ExternalOrganization struct {
	ID   string
	Name string
}

type OrganizationGetter interface {
	GetOrganization(ctx context.Context, orgID uuid.UUID) (*api.Organization, api.Status)
}

type OrganizationValidator interface {
	ValidateOrganization(ctx context.Context, resource string, action string) error
}

type OrganizationExistsValidator struct {
	orgResolver *org.Resolver
}

func NewOrganizationExistsValidator(orgResolver *org.Resolver) *OrganizationExistsValidator {
	return &OrganizationExistsValidator{
		orgResolver: orgResolver,
	}
}

func (v *OrganizationExistsValidator) ValidateOrganization(ctx context.Context, resource string, action string) error {
	// Skip org validation for listing all organizations - users should be able to
	// list organizations they belong to without requiring a specific org context
	if resource == "organizations" && action == "list" {
		return nil
	}

	orgID, ok := util.GetOrgIdFromContext(ctx)
	if !ok {
		return fmt.Errorf("no org id in context")
	}

	exists, err := v.orgResolver.ValidateAccess(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to validate organization: %w", err)
	}
	if !exists {
		return fmt.Errorf("organization not found: %s", orgID)
	}

	return nil
}

func GetIdentity(ctx context.Context) (*Identity, error) {
	identityVal := ctx.Value(IdentityCtxKey)
	if identityVal == nil {
		return nil, fmt.Errorf("failed to get identity from context")
	}
	identity, ok := identityVal.(*Identity)
	if !ok {
		return nil, fmt.Errorf("incorrect type of identity in context")
	}
	return identity, nil
}

func ExtractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get(AuthHeader)
	if authHeader == "" {
		return "", fmt.Errorf("empty %s header", AuthHeader)
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		return "", fmt.Errorf("invalid %s header", AuthHeader)
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", fmt.Errorf("invalid token")
	}
	return token, nil
}
