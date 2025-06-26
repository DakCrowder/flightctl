package model

import (
	"time"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/google/uuid"
)

type Organization struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	// Human readable name shown to users
	DisplayName string `gorm:"not null"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewOrganizationFromApiResource(resource *api.Organization) (*Organization, error) {
	return &Organization{
		ID:          *resource.Id,
		DisplayName: *resource.DisplayName,
	}, nil
}

func (o *Organization) ToApiResource(opts ...APIResourceOption) (*api.Organization, error) {
	if o == nil {
		return &api.Organization{}, nil
	}

	return &api.Organization{
		Id:          &o.ID,
		DisplayName: &o.DisplayName,
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
