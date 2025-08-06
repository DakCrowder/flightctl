package providers

import (
	"context"
)

type JWTProvider struct {
}

func NewJWTProvider() *JWTProvider {
	return &JWTProvider{}
}

func (p *JWTProvider) GetUserOrgs(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (p *JWTProvider) HasAccess(ctx context.Context, externalOrgID string) (bool, error) {
	return false, nil
}
