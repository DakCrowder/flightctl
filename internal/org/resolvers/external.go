package resolvers

import (
	"context"
	"slices"
	"time"

	"github.com/flightctl/flightctl/internal/auth/common"
	"github.com/flightctl/flightctl/internal/org/providers"
	"github.com/flightctl/flightctl/internal/store/model"
	"github.com/google/uuid"
	"github.com/jellydator/ttlcache/v3"
	"github.com/sirupsen/logrus"
)

// ExternalResolver caches organization ID validation.
type ExternalResolver struct {
	store               OrgStore
	cache               *ttlcache.Cache[uuid.UUID, *model.Organization]
	ttl                 time.Duration
	externalOrgProvider providers.ExternalOrganizationProvider
	log                 logrus.FieldLogger
}

// NewExternalResolver constructs a new resolver. A TTL of zero disables expiration.
func NewExternalResolver(s OrgStore, ttl time.Duration, externalOrgProvider providers.ExternalOrganizationProvider, log logrus.FieldLogger) *ExternalResolver {
	opts := []ttlcache.Option[uuid.UUID, *model.Organization]{}
	if ttl > 0 {
		opts = append(opts, ttlcache.WithTTL[uuid.UUID, *model.Organization](ttl))
	}
	c := ttlcache.New(opts...)
	go c.Start()
	return &ExternalResolver{
		store:               s,
		cache:               c,
		ttl:                 ttl,
		externalOrgProvider: externalOrgProvider,
		log:                 log,
	}
}

func (r *ExternalResolver) EnsureExists(ctx context.Context, id uuid.UUID) error {
	_, err := r.getOrg(ctx, id)
	return err
}

// getOrg fetches the organization for the given ID. It caches lookups according
// to the configured TTL. Failed lookups are not cached, ensuring that newly
// created organizations are immediately accessible.
func (r *ExternalResolver) getOrg(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	if item := r.cache.Get(id); item != nil {
		return item.Value(), nil
	}
	// Cache miss – query the store.
	org, err := r.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Use configured TTL or disable expiration if TTL is not positive
	cacheTTL := ttlcache.NoTTL
	if r.ttl > 0 {
		cacheTTL = r.ttl
	}
	r.cache.Set(id, org, cacheTTL)
	return org, nil
}

// Close stops the cache and releases resources. Should be called when the resolver
// is no longer needed to prevent goroutine leaks.
func (r *ExternalResolver) Close() {
	r.cache.Stop()
}

func (r *ExternalResolver) IsMemberOf(ctx context.Context, identity common.Identity, id uuid.UUID) (bool, error) {
	org, err := r.getOrg(ctx, id)
	if err != nil {
		return false, err
	}

	return r.externalOrgProvider.IsMemberOf(ctx, identity, org.ExternalID)
}

func (r *ExternalResolver) GetUserOrganizations(ctx context.Context, identity common.Identity) ([]*model.Organization, error) {
	externalOrgs, err := r.externalOrgProvider.GetUserOrganizations(ctx, identity)
	if err != nil {
		return nil, err
	}

	if len(externalOrgs) == 0 {
		return []*model.Organization{}, nil
	}

	externalOrgIDs := make([]string, 0, len(externalOrgs))
	for _, org := range externalOrgs {
		externalOrgIDs = append(externalOrgIDs, org.ID)
	}

	organizations, err := r.store.ListByExternalIDs(ctx, externalOrgIDs)
	if err != nil {
		return nil, err
	}

	// Create a list of existing external org IDs in our database
	existingExternalOrgIDs := make([]string, 0, len(organizations))
	for _, org := range organizations {
		existingExternalOrgIDs = append(existingExternalOrgIDs, org.ExternalID)
	}

	// Find external org IDs that are not yet in our database
	newExternalOrgIDs := make([]string, 0)
	for _, externalOrgID := range externalOrgIDs {
		if !slices.Contains(existingExternalOrgIDs, externalOrgID) {
			newExternalOrgIDs = append(newExternalOrgIDs, externalOrgID)
		}
	}

	if len(newExternalOrgIDs) > 0 {
		// Create a map for quick lookup of external org names by ID
		externalOrgMap := make(map[string]string, len(externalOrgs))
		for _, org := range externalOrgs {
			externalOrgMap[org.ID] = org.Name
		}

		newOrgs := make([]*model.Organization, len(newExternalOrgIDs))
		for i, orgID := range newExternalOrgIDs {
			id := uuid.New()
			displayName := externalOrgMap[orgID]
			if displayName == "" {
				// Fallback to UUID string if name is not found
				r.log.Warnf("external org name not found, using UUID %s", id.String())
				displayName = id.String()
			}
			newOrgs[i] = &model.Organization{
				ID:          id,
				ExternalID:  orgID,
				DisplayName: displayName,
			}
		}

		newOrgs, err = r.store.CreateMany(ctx, newOrgs)
		if err != nil {
			return nil, err
		}
		organizations = append(organizations, newOrgs...)
	}

	return organizations, nil
}
