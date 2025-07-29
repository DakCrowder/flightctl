package service

import (
	"context"
	"fmt"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/flightctl/flightctl/internal/auth/common"
	"github.com/google/uuid"
)

var organizationApiVersion = fmt.Sprintf("%s/%s", api.APIGroup, api.OrganizationAPIVersion)

func (h *ServiceHandler) ListOrganizations(ctx context.Context) (*api.OrganizationList, api.Status) {
	// TODO Change identity to be a different format
	identity, err := common.GetIdentity(ctx)
	if err != nil {
		return nil, StoreErrorToApiStatus(err, false, api.OrganizationKind, nil)
	}

	orgs, err := h.store.Organization().ListAndCreateMissing(ctx, identity.Organizations)
	status := StoreErrorToApiStatus(err, false, api.OrganizationKind, nil)
	if err != nil {
		return nil, status
	}

	apiOrgs := make([]api.Organization, len(orgs))
	for i, org := range orgs {
		name := org.ID.String()
		apiOrgs[i] = api.Organization{
			ApiVersion: organizationApiVersion,
			Kind:       api.OrganizationKind,
			Metadata:   api.ObjectMeta{Name: &name},
			Spec: &api.OrganizationSpec{
				ExternalId:  &org.ExternalID,
				DisplayName: &org.DisplayName,
			},
		}
	}

	return &api.OrganizationList{
		Items:      apiOrgs,
		ApiVersion: organizationApiVersion,
		Kind:       api.OrganizationListKind,
		Metadata:   api.ListMeta{},
	}, status
}

func (h *ServiceHandler) GetOrganization(ctx context.Context, orgID uuid.UUID) (*api.Organization, api.Status) {
	org, err := h.store.Organization().GetByID(ctx, orgID)
	status := StoreErrorToApiStatus(err, false, api.OrganizationKind, nil)
	if err != nil {
		return nil, status
	}
	name := org.ID.String()
	return &api.Organization{
		ApiVersion: organizationApiVersion,
		Kind:       api.OrganizationKind,
		Metadata:   api.ObjectMeta{Name: &name},
		Spec: &api.OrganizationSpec{
			ExternalId:  &org.ExternalID,
			DisplayName: &org.DisplayName,
		},
	}, status
}
