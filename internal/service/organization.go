package service

import (
	"context"

	api "github.com/flightctl/flightctl/api/v1alpha1"
)

func (h *ServiceHandler) ListOrganizations(ctx context.Context) (*api.OrganizationList, api.Status) {
	result, err := h.store.Organization().List(ctx)
	if err != nil {
		return nil, api.StatusInternalServerError(err.Error())
	}

	return result, api.StatusOK()
}
