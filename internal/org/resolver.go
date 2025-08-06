package org

import (
	"context"
	"time"

	"github.com/flightctl/flightctl/internal/store/model"
	"github.com/google/uuid"
	"github.com/jellydator/ttlcache/v3"
)

// ExternalOrgProvider fetches external organization IDs that a user has access to.
// Different implementations can fetch from JWTs, external APIs, etc.
type ExternalOrgProvider interface {
	// GetUserOrgs returns external org IDs the user has access to
	GetUserOrgs(ctx context.Context) ([]string, error)

	// HasAccess checks if user has access to a specific external org
	HasAccess(ctx context.Context, externalOrgID string) (bool, error)
}

// OrgLookup retrieves an organization by ID.
type OrgLookup interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error)
}

// Resolver caches organization ID validation.
type Resolver struct {
	store               OrgLookup
	externalOrgProvider ExternalOrgProvider
	internalCache       *ttlcache.Cache[uuid.UUID, *model.Organization]
	externalCache       *ttlcache.Cache[string, *model.Organization]
	ttl                 time.Duration
}

type ResolverConfig struct {
	Store            OrgLookup
	ExternalProvider ExternalOrgProvider
	CacheTTL         time.Duration
}

// NewResolver constructs a new resolver. A TTL of zero disables expiration.
func NewResolver(cfg ResolverConfig) *Resolver {
	opts := []ttlcache.Option[uuid.UUID, *model.Organization]{}
	extOpts := []ttlcache.Option[string, *model.Organization]{}

	if cfg.CacheTTL > 0 {
		opts = append(opts, ttlcache.WithTTL[uuid.UUID, *model.Organization](cfg.CacheTTL))
		extOpts = append(extOpts, ttlcache.WithTTL[string, *model.Organization](cfg.CacheTTL))
	}

	cache := ttlcache.New(opts...)
	externalCache := ttlcache.New(extOpts...)

	go cache.Start()
	go externalCache.Start()

	return &Resolver{
		store:               cfg.Store,
		externalOrgProvider: cfg.ExternalProvider,
		internalCache:       cache,
		externalCache:       externalCache,
		ttl:                 cfg.CacheTTL,
	}
}

// EnsureExists checks that the given organization ID exists. It caches positive
// look-ups according to the configured TTL. Failed validations are not cached,
// ensuring that newly created organizations are immediately accessible.
func (r *Resolver) EnsureExists(ctx context.Context, id uuid.UUID) error {
	_, err := r.getByID(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (r *Resolver) GetExternalID(ctx context.Context, id uuid.UUID) (string, error) {
	org, err := r.store.GetByID(ctx, id)
	if err != nil {
		return "", err
	}
	return org.ExternalID, nil
}

func (r *Resolver) ValidateAccess(ctx context.Context, orgID uuid.UUID) (bool, error) {
	// Fetch org from internal store to verify it exists
	org, err := r.getByID(ctx, orgID)
	if err != nil {
		return false, err
	}

	// Check external org provider for access
	hasAccess, err := r.externalOrgProvider.HasAccess(ctx, org.ExternalID)
	if err != nil {
		return false, err
	}
	return hasAccess, nil
}

func (r *Resolver) getByID(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	if item := r.internalCache.Get(id); item != nil {
		return item.Value(), nil
	}

	// Cache miss, fetch from store
	org, err := r.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	r.cacheOrg(org)
	return org, nil
}

// Caches an org in both caches
func (r *Resolver) cacheOrg(org *model.Organization) {
	cacheTTL := ttlcache.NoTTL
	if r.ttl > 0 {
		cacheTTL = r.ttl
	}
	r.internalCache.Set(org.ID, org, cacheTTL)
	r.externalCache.Set(org.ExternalID, org, cacheTTL)
}

// Close stops the cache and releases resources. Should be called when the resolver
// is no longer needed to prevent goroutine leaks.
func (r *Resolver) Close() {
	r.internalCache.Stop()
}
