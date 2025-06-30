package transport

import (
	"encoding/json"
	"net/http"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/oapi-codegen/runtime/types"
)

// (POST /api/v1/organizations/{orgID}/repositories)
func (h *TransportHandler) CreateRepository(w http.ResponseWriter, r *http.Request, orgID types.UUID) {
	var repository api.Repository
	if err := json.NewDecoder(r.Body).Decode(&repository); err != nil {
		SetParseFailureResponse(w, err)
		return
	}

	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.CreateRepository(r.Context(), orgUUID, repository)
	SetResponse(w, body, status)
}

// (GET /api/v1/organizations/{orgID}/repositories)
func (h *TransportHandler) ListRepositories(w http.ResponseWriter, r *http.Request, orgID types.UUID, params api.ListRepositoriesParams) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.ListRepositories(r.Context(), orgUUID, params)
	SetResponse(w, body, status)
}

// (GET /api/v1/organizations/{orgID}/repositories/{name})
func (h *TransportHandler) GetRepository(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.GetRepository(r.Context(), orgUUID, name)
	SetResponse(w, body, status)
}

// (PUT /api/v1/organizations/{orgID}/repositories/{name})
func (h *TransportHandler) ReplaceRepository(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	var repository api.Repository
	if err := json.NewDecoder(r.Body).Decode(&repository); err != nil {
		SetParseFailureResponse(w, err)
		return
	}

	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.ReplaceRepository(r.Context(), orgUUID, name, repository)
	SetResponse(w, body, status)
}

// (DELETE /api/v1/organizations/{orgID}/repositories/{name})
func (h *TransportHandler) DeleteRepository(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	status := h.serviceHandler.DeleteRepository(r.Context(), orgUUID, name)
	SetResponse(w, nil, status)
}

// (PATCH /api/v1/organizations/{orgID}/repositories/{name})
func (h *TransportHandler) PatchRepository(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	var patch api.PatchRequest
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		SetParseFailureResponse(w, err)
		return
	}

	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.PatchRepository(r.Context(), orgUUID, name, patch)
	SetResponse(w, body, status)
}
