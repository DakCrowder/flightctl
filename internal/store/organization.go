package store

import (
	"context"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/flightctl/flightctl/internal/store/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Organization interface {
	InitialMigration(ctx context.Context) error

	List(ctx context.Context) (*api.OrganizationList, error)
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
