package transport

import (
	"encoding/json"
	"net/http"

	api "github.com/flightctl/flightctl/api/v1alpha1"
)

// (GET /api/v1/users/me/organizations)
func (h *TransportHandler) ListUserOrganizations(w http.ResponseWriter, r *http.Request) {
	body, status := h.serviceHandler.ListUserOrganizations(r.Context())
	SetResponse(w, body, status)
}

// (POST /api/v1/organizations)
func (h *TransportHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var organization api.Organization
	if err := json.NewDecoder(r.Body).Decode(&organization); err != nil {
		SetParseFailureResponse(w, err)
		return
	}

	body, status := h.serviceHandler.CreateOrganization(r.Context(), organization)
	SetResponse(w, body, status)
}

// (PUT /api/v1/organizations/{orgID})
func (h *TransportHandler) ReplaceOrganization(w http.ResponseWriter, r *http.Request, orgID string) {
	var organization api.Organization
	if err := json.NewDecoder(r.Body).Decode(&organization); err != nil {
		SetParseFailureResponse(w, err)
		return
	}

	body, status := h.serviceHandler.ReplaceOrganization(r.Context(), orgID, organization)
	SetResponse(w, body, status)
}

// (GET /api/v1/organizations)
func (h *TransportHandler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	body, status := h.serviceHandler.ListOrganizations(r.Context())
	SetResponse(w, body, status)
}
