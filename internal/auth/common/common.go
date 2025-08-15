package common

import (
	"context"
	"fmt"
	"net/http"
	"strings"
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
	Type                string
	Url                 string
	OrganizationsConfig AuthOrganizationsConfig
}

type AuthOrganizationsConfig struct {
	Enabled bool
}

type Identity interface {
	GetUsername() string
	SetUsername(username string)
	GetUID() string
	SetUID(uID string)
	GetGroups() []string
	SetGroups(groups []string)
}

type UserIdentity struct {
	username string
	uID      string
	groups   []string
}

func NewUserIdentity(username string, uID string, groups []string) *UserIdentity {
	return &UserIdentity{
		username: username,
		uID:      uID,
		groups:   groups,
	}
}

func (i *UserIdentity) GetUsername() string {
	return i.username
}

func (i *UserIdentity) SetUsername(username string) {
	i.username = username
}

func (i *UserIdentity) GetUID() string {
	return i.uID
}

func (i *UserIdentity) SetUID(uID string) {
	i.uID = uID
}

func (i *UserIdentity) GetGroups() []string {
	return i.groups
}

func (i *UserIdentity) SetGroups(groups []string) {
	i.groups = groups
}

func GetIdentity(ctx context.Context) (Identity, error) {
	identityVal := ctx.Value(IdentityCtxKey)
	if identityVal == nil {
		return nil, fmt.Errorf("failed to get identity from context")
	}
	identity, ok := identityVal.(*UserIdentity)
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
