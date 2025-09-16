package resolvers

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/flightctl/flightctl/internal/auth/common"
	"github.com/flightctl/flightctl/internal/config"
	"github.com/flightctl/flightctl/internal/org/cache"
	"github.com/flightctl/flightctl/internal/org/providers"
	"github.com/flightctl/flightctl/internal/store/model"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type OrgStore interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error)
	ListByExternalIDs(ctx context.Context, externalIDs []string) ([]*model.Organization, error)
	UpsertMany(ctx context.Context, orgs []*model.Organization) ([]*model.Organization, error)
}

type Resolver interface {
	EnsureExists(ctx context.Context, id uuid.UUID) error
	IsMemberOf(ctx context.Context, identity common.Identity, id uuid.UUID) (bool, error)
	GetUserOrganizations(ctx context.Context, identity common.Identity) ([]*model.Organization, error)
}

type BuildResolverOptions struct {
	Config *config.Config
	Store  OrgStore
	Log    logrus.FieldLogger
	Cache  cache.OrganizationCache
}

func BuildResolver(opts BuildResolverOptions) (Resolver, error) {
	if opts.Config != nil && opts.Config.Auth != nil && opts.Config.Organizations != nil && opts.Config.Organizations.Enabled {
		if opts.Config.Auth.OIDC != nil {
			return buildOIDCResolver(opts), nil
		} else if opts.Config.Auth.AAP != nil {
			return buildAAPResolver(opts)
		} else if opts.Config.Auth.K8s != nil {
			opts.Log.Warn("K8s organizations are not supported yet, falling back to default resolver")
		} else {
			opts.Log.Warn("No auth provider properly configured, falling back to default resolver")
		}
	}

	return NewDefaultResolver(opts.Store, opts.Cache), nil
}

func buildOIDCResolver(opts BuildResolverOptions) Resolver {
	return NewExternalResolver(opts.Store, opts.Cache, &providers.ClaimsProvider{}, opts.Log)
}

func buildAAPResolver(opts BuildResolverOptions) (Resolver, error) {
	provider, err := providers.NewAAPOrganizationProvider(opts.Config.Auth.AAP.ApiUrl, &tls.Config{
		InsecureSkipVerify: opts.Config.Auth.InsecureSkipTlsVerify,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AAP organization provider: %w", err)
	}
	return NewExternalResolver(opts.Store, opts.Cache, provider, opts.Log), nil
}
