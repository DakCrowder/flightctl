package service

import (
	"context"

	api "github.com/flightctl/flightctl/api/v1alpha1"
)

func (h *ServiceHandler) ListUserOrganizations(ctx context.Context) (*api.OrganizationList, api.Status) {
	result, err := h.store.Organization().List(ctx)
	if err != nil {
		return nil, api.StatusInternalServerError(err.Error())
	}

	// Add default displayName to organizations
	for i := range result.Items {
		if result.Items[i].DisplayName == nil || *result.Items[i].DisplayName == "" {
			defaultName := "Default Organization"
			result.Items[i].DisplayName = &defaultName
		}
	}

	return result, api.StatusOK()
}
