package service

import (
	"context"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/google/uuid"
)

func (h *ServiceHandler) ListOrganizations(ctx context.Context) (*api.OrganizationList, api.Status) {
	result, err := h.store.Organization().List(ctx)
	if err != nil {
		return nil, api.StatusInternalServerError(err.Error())
	}

	return result, api.StatusOK()
}

func (h *ServiceHandler) ListUserOrganizations(ctx context.Context) (*api.OrganizationList, api.Status) {
	result, err := h.store.Organization().List(ctx)
	if err != nil {
		return nil, api.StatusInternalServerError(err.Error())
	}

	return result, api.StatusOK()
}

func (h *ServiceHandler) CreateOrganization(ctx context.Context, org api.Organization) (*api.Organization, api.Status) {
	result, err := h.store.Organization().Create(ctx, &org)
	if err != nil {
		return nil, api.StatusInternalServerError(err.Error())
	}

	// TODO create event

	return result, api.StatusOK()
}

func (h *ServiceHandler) ReplaceOrganization(ctx context.Context, orgID string, org api.Organization) (*api.Organization, api.Status) {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, api.StatusBadRequest(err.Error())
	}

	result, _, err := h.store.Organization().Update(ctx, orgUUID, &org)
	if err != nil {
		return nil, api.StatusInternalServerError(err.Error())
	}

	// TODO create event

	return result, api.StatusOK()
}
