package store

import (
	"context"

	"github.com/flightctl/flightctl/internal/auth/common"
	"github.com/flightctl/flightctl/internal/flterrors"
	"github.com/flightctl/flightctl/internal/store/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Organization interface {
	InitialMigration(ctx context.Context) error

	Create(ctx context.Context, org *model.Organization) (*model.Organization, error)
	List(ctx context.Context) ([]*model.Organization, error)
	ListAndCreateMissing(ctx context.Context, orgs []common.ExternalOrganization) ([]*model.Organization, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error)
}

type OrganizationStore struct {
	dbHandler *gorm.DB
}

// Ensure OrganizationStore implements the Organization interface
var _ Organization = (*OrganizationStore)(nil)

func NewOrganization(db *gorm.DB) Organization {
	return &OrganizationStore{dbHandler: db}
}

func (s *OrganizationStore) getDB(ctx context.Context) *gorm.DB {
	return s.dbHandler.WithContext(ctx)
}

func (s *OrganizationStore) InitialMigration(ctx context.Context) error {
	db := s.getDB(ctx)
	if err := db.AutoMigrate(&model.Organization{}); err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.Organization{}).Count(&count).Error; err != nil {
			return err
		}

		// If there are no organizations, create a default one
		if count == 0 {
			if err := tx.Create(&model.Organization{
				ID:          NullOrgId,
				DisplayName: "Default",
			}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *OrganizationStore) Create(ctx context.Context, org *model.Organization) (*model.Organization, error) {
	db := s.getDB(ctx)

	if org.ID == uuid.Nil {
		org.ID = uuid.New()
	}

	if err := db.Create(org).Error; err != nil {
		return nil, err
	}

	return org, nil
}

func (s *OrganizationStore) GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	db := s.getDB(ctx)

	var org model.Organization
	result := db.Where("id = ?", id).Take(&org)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, flterrors.ErrResourceNotFound
		}
		return nil, result.Error
	}

	return &org, nil
}

func (s *OrganizationStore) List(ctx context.Context) ([]*model.Organization, error) {
	db := s.getDB(ctx)

	var orgs []*model.Organization
	if err := db.Find(&orgs).Error; err != nil {
		return nil, err
	}

	return orgs, nil
}

// ListAndCreateMissing lists existing orgs and creates any that are missing from the provided list
func (s *OrganizationStore) ListAndCreateMissing(ctx context.Context, orgs []common.ExternalOrganization) ([]*model.Organization, error) {
	if len(orgs) == 0 {
		return []*model.Organization{}, nil
	}

	db := s.getDB(ctx)

	// Step 1: Find all existing organizations using a single 'IN' query.
	externalIDs := make([]string, len(orgs))
	for i, org := range orgs {
		externalIDs[i] = org.ID
	}
	var foundOrgs []*model.Organization
	if err := db.Where("external_id IN ?", externalIDs).Find(&foundOrgs).Error; err != nil {
		return nil, err
	}

	// Step 2: Use a map to efficiently identify which external IDs are missing.
	idsToFind := make(map[string]bool)
	for _, id := range externalIDs {
		idsToFind[id] = true
	}

	for _, org := range foundOrgs {
		// This ID was found, so remove it from the map of IDs we need to create.
		delete(idsToFind, org.ExternalID)
	}

	// Step 3: Prepare and bulk-create the missing organizations.
	// TODO populate display name appropriately
	var createdOrgs []*model.Organization
	if len(idsToFind) > 0 {
		var orgsToCreate []*model.Organization
		for id := range idsToFind {
			orgsToCreate = append(orgsToCreate, &model.Organization{
				ID:         uuid.New(),
				ExternalID: id,
				// A reasonable default is to set the DisplayName to the ExternalID.
				// You can adjust this based on your business logic.
				DisplayName: id,
			})
		}

		// GORM's Create with a slice performs a bulk insert.
		// It also populates the slice with primary keys and timestamps from the DB.
		if err := db.Create(&orgsToCreate).Error; err != nil {
			return nil, err
		}
		createdOrgs = orgsToCreate
	}

	// Step 4: Combine the found and newly created records and return.
	allOrgs := append(foundOrgs, createdOrgs...)
	return allOrgs, nil
}
