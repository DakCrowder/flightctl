package service

import (
	"context"
	"fmt"
	"slices"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/flightctl/flightctl/internal/auth/common"
	"github.com/flightctl/flightctl/internal/store/model"
	"github.com/flightctl/flightctl/internal/util"
	"github.com/google/uuid"
)

var organizationApiVersion = fmt.Sprintf("%s/%s", api.APIGroup, api.OrganizationAPIVersion)

func (h *ServiceHandler) ListOrganizations(ctx context.Context) (*api.OrganizationList, api.Status) {

	var orgs []*model.Organization
	var err error
	if util.IsInternalRequest(ctx) {
		orgs, err = h.listSystemOrganizations(ctx)
	} else {
		orgs, err = h.listUserScopedOrganizations(ctx)
	}

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

func (h *ServiceHandler) listSystemOrganizations(ctx context.Context) ([]*model.Organization, error) {
	return h.store.Organization().List(ctx)
}

func (h *ServiceHandler) listUserScopedOrganizations(ctx context.Context) ([]*model.Organization, error) {
	// TODO Change identity to be a different format
	identity, err := common.GetIdentity(ctx)
	if err != nil {
		return nil, err
	}

	externalOrgIDs := make([]string, len(identity.Organizations))
	for i, org := range identity.Organizations {
		externalOrgIDs[i] = org.ID
	}

	organizations, err := h.store.Organization().ListByExternalIDs(ctx, externalOrgIDs)
	if err != nil {
		return nil, err
	}

	// TODO this is all quite inefficient, once other types are reconciled revist
	newExternalOrgIDs := make([]string, 0)
	for _, org := range organizations {
		// if the external orgid in externalOrgIDs is not present in organizations, add it to newExternalOrgIDs
		if !slices.Contains(externalOrgIDs, org.ExternalID) {
			newExternalOrgIDs = append(newExternalOrgIDs, org.ExternalID)
		}
	}

	if len(newExternalOrgIDs) > 0 {
		newOrgs := make([]*model.Organization, len(newExternalOrgIDs))
		for i, orgID := range newExternalOrgIDs {
			// TODO populate name
			id := uuid.New()
			newOrgs[i] = &model.Organization{
				ID:          id,
				ExternalID:  orgID,
				DisplayName: id.String(),
			}
		}

		newOrgs, err = h.store.Organization().CreateMany(ctx, newOrgs)
		if err != nil {
			return nil, err
		}
		organizations = append(organizations, newOrgs...)
	}

	return organizations, nil
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
