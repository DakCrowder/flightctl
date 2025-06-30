package store

import (
	"context"
	"errors"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/flightctl/flightctl/internal/store/model"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Organization interface {
	InitialMigration(ctx context.Context) error

	Create(ctx context.Context, resource *api.Organization) (*api.Organization, error)
	Update(ctx context.Context, orgID uuid.UUID, resource *api.Organization) (*api.Organization, api.ResourceUpdatedDetails, error)
	List(ctx context.Context) (*api.OrganizationList, error)
	Get(ctx context.Context, orgID uuid.UUID) (*api.Organization, error)
}

type OrganizationStore struct {
	dbHandler *gorm.DB
	log       logrus.FieldLogger
}

// Make sure we conform to Event interface
var _ Organization = (*OrganizationStore)(nil)

func NewOrganization(db *gorm.DB, log logrus.FieldLogger) Organization {
	return &OrganizationStore{dbHandler: db, log: log}
}

func (s *OrganizationStore) getDB(ctx context.Context) *gorm.DB {
	return s.dbHandler.WithContext(ctx)
}

func (s *OrganizationStore) InitialMigration(ctx context.Context) error {
	db := s.getDB(ctx)

	if err := db.AutoMigrate(&model.Organization{}); err != nil {
		return err
	}

	db.Create(&model.Organization{
		ID:          NullOrgId,
		DisplayName: "Default",
	})

	return nil
}

func (s *OrganizationStore) Create(ctx context.Context, resource *api.Organization) (*api.Organization, error) {
	if resource == nil {
		return nil, errors.New("resource cannot be nil")
	}

	organization := &model.Organization{
		ID:          uuid.New(),
		DisplayName: *resource.DisplayName,
	}

	db := s.getDB(ctx)
	result := db.Create(organization)
	if result.Error != nil {
		return nil, ErrorFromGormError(result.Error)
	}

	return organization.ToApiResource()
}

func (s *OrganizationStore) Update(ctx context.Context, orgID uuid.UUID, resource *api.Organization) (*api.Organization, api.ResourceUpdatedDetails, error) {
	if resource == nil {
		return nil, api.ResourceUpdatedDetails{}, errors.New("resource cannot be nil")
	}

	var existingOrganization model.Organization
	db := s.getDB(ctx)
	result := db.Where("id = ?", orgID).First(&existingOrganization)
	if result.Error != nil {
		return nil, api.ResourceUpdatedDetails{}, ErrorFromGormError(result.Error)
	}

	// Currently only the display name can be updated
	var updatedFields []api.ResourceUpdatedDetailsUpdatedFields
	if existingOrganization.DisplayName != *resource.DisplayName {
		updatedFields = append(updatedFields, "displayName")
		existingOrganization.DisplayName = *resource.DisplayName
	} else {
		// NOOP - nothing to update
		return resource, api.ResourceUpdatedDetails{}, nil
	}

	result = db.Save(&existingOrganization)
	if result.Error != nil {
		return nil, api.ResourceUpdatedDetails{}, ErrorFromGormError(result.Error)
	}

	updatedResource, err := existingOrganization.ToApiResource()
	if err != nil {
		return nil, api.ResourceUpdatedDetails{}, err
	}

	details := api.ResourceUpdatedDetails{
		UpdatedFields: updatedFields,
	}

	return updatedResource, details, nil
}

func (s *OrganizationStore) List(ctx context.Context) (*api.OrganizationList, error) {
	var organizationsList []model.Organization
	query := s.getDB(ctx).Model(&model.Organization{}).Order("created_at DESC")

	result := query.Find(&organizationsList)
	if result.Error != nil {
		return nil, ErrorFromGormError(result.Error)
	}

	apiOrganizationList, err := model.OrganizationsToApiResource(organizationsList)
	if err != nil {
		return nil, err
	}

	return apiOrganizationList, nil
}

func (s *OrganizationStore) Get(ctx context.Context, orgID uuid.UUID) (*api.Organization, error) {
	var organization model.Organization
	db := s.getDB(ctx)

	result := db.Where("id = ?", orgID).First(&organization)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrorFromGormError(result.Error)
		}
		return nil, ErrorFromGormError(result.Error)
	}

	return organization.ToApiResource()
}
