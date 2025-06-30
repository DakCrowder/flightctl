package transport

import (
	"net/http"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/oapi-codegen/runtime/types"
)

// (GET /api/v1/organizations/{orgID}/fleets/{fleet}/templateversions)
func (h *TransportHandler) ListTemplateVersions(w http.ResponseWriter, r *http.Request, orgID types.UUID, fleet string, params api.ListTemplateVersionsParams) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.ListTemplateVersions(r.Context(), orgUUID, fleet, params)
	SetResponse(w, body, status)
}

// (GET /api/v1/organizations/{orgID}/fleets/{fleet}/templateversions/{name})
func (h *TransportHandler) GetTemplateVersion(w http.ResponseWriter, r *http.Request, orgID types.UUID, fleet string, name string) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.GetTemplateVersion(r.Context(), orgUUID, fleet, name)
	SetResponse(w, body, status)
}

// (DELETE /api/v1/organizations/{orgID}/fleets/{fleet}/templateversions/{name})
func (h *TransportHandler) DeleteTemplateVersion(w http.ResponseWriter, r *http.Request, orgID types.UUID, fleet string, name string) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	status := h.serviceHandler.DeleteTemplateVersion(r.Context(), orgUUID, fleet, name)
	SetResponse(w, nil, status)
}
