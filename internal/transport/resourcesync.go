package transport

import (
	"encoding/json"
	"net/http"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/oapi-codegen/runtime/types"
)

// (POST /api/v1/organizations/{orgID}/resourcesyncs)
func (h *TransportHandler) CreateResourceSync(w http.ResponseWriter, r *http.Request, orgID types.UUID) {
	var resourceSync api.ResourceSync
	if err := json.NewDecoder(r.Body).Decode(&resourceSync); err != nil {
		SetParseFailureResponse(w, err)
		return
	}

	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.CreateResourceSync(r.Context(), orgUUID, resourceSync)
	SetResponse(w, body, status)
}

// (GET /api/v1/organizations/{orgID}/resourcesyncs)
func (h *TransportHandler) ListResourceSyncs(w http.ResponseWriter, r *http.Request, orgID types.UUID, params api.ListResourceSyncsParams) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.ListResourceSyncs(r.Context(), orgUUID, params)
	SetResponse(w, body, status)
}

// (GET /api/v1/organizations/{orgID}/resourcesyncs/{name})
func (h *TransportHandler) GetResourceSync(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.GetResourceSync(r.Context(), orgUUID, name)
	SetResponse(w, body, status)
}

// (PUT /api/v1/organizations/{orgID}/resourcesyncs/{name})
func (h *TransportHandler) ReplaceResourceSync(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	var resourceSync api.ResourceSync
	if err := json.NewDecoder(r.Body).Decode(&resourceSync); err != nil {
		SetParseFailureResponse(w, err)
		return
	}

	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.ReplaceResourceSync(r.Context(), orgUUID, name, resourceSync)
	SetResponse(w, body, status)
}

// (DELETE /api/v1/organizations/{orgID}/resourcesyncs/{name})
func (h *TransportHandler) DeleteResourceSync(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	status := h.serviceHandler.DeleteResourceSync(r.Context(), orgUUID, name)
	SetResponse(w, nil, status)
}

// (PATCH /api/v1/organizations/{orgID}/resourcesyncs/{name})
func (h *TransportHandler) PatchResourceSync(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
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

	body, status := h.serviceHandler.PatchResourceSync(r.Context(), orgUUID, name, patch)
	SetResponse(w, body, status)
}
