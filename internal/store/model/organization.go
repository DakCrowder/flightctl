package model

import (
	"time"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/google/uuid"
)

type Organization struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	// Unique external identifier for the organization
	ExternalID string `gorm:"uniqueIndex"`

	// Source of the external ID (e.g., "aap", "oidc")
	ExternalIDSource string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewOrganizationFromApiResource(resource *api.Organization) (*Organization, error) {
	org := &Organization{
		ID: *resource.Id,
	}
	
	// ExternalID and ExternalIDSource would need to be set from the API resource
	// if they were added to the API spec
	
	return org, nil
}

func (o *Organization) ToApiResource(opts ...APIResourceOption) (*api.Organization, error) {
	if o == nil {
		return &api.Organization{}, nil
	}

	return &api.Organization{
		Id: &o.ID,
		// ExternalID and ExternalIDSource would need to be added to the API resource
		// if they were added to the API spec
	}, nil
}

func OrganizationsToApiResource(organizations []Organization) (*api.OrganizationList, error) {
	orgs := make([]api.Organization, len(organizations))
	for i, organization := range organizations {
		org, err := organization.ToApiResource()
		if err != nil {
			return nil, err
		}
		orgs[i] = *org
	}

	return &api.OrganizationList{
		Items: orgs,
	}, nil
}
