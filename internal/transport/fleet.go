package transport

import (
	"encoding/json"
	"net/http"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/oapi-codegen/runtime/types"
)

// (POST /api/v1/organizations/{orgID}/fleets)
func (h *TransportHandler) CreateFleet(w http.ResponseWriter, r *http.Request, orgID types.UUID) {
	var fleet api.Fleet
	if err := json.NewDecoder(r.Body).Decode(&fleet); err != nil {
		SetParseFailureResponse(w, err)
		return
	}

	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.CreateFleet(r.Context(), orgUUID, fleet)
	SetResponse(w, body, status)
}

// (GET /api/v1/organizations/{orgID}/fleets)
func (h *TransportHandler) ListFleets(w http.ResponseWriter, r *http.Request, orgID types.UUID, params api.ListFleetsParams) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.ListFleets(r.Context(), orgUUID, params)
	SetResponse(w, body, status)
}

// (GET /api/v1/organizations/{orgID}/fleets/{name})
func (h *TransportHandler) GetFleet(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string, params api.GetFleetParams) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.GetFleet(r.Context(), orgUUID, name, params)
	SetResponse(w, body, status)
}

// (PUT /api/v1/organizations/{orgID}/fleets/{name})
func (h *TransportHandler) ReplaceFleet(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	var fleet api.Fleet
	if err := json.NewDecoder(r.Body).Decode(&fleet); err != nil {
		SetParseFailureResponse(w, err)
		return
	}

	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.ReplaceFleet(r.Context(), orgUUID, name, fleet)
	SetResponse(w, body, status)
}

// (DELETE /api/v1/organizations/{orgID}/fleets/{name})
func (h *TransportHandler) DeleteFleet(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	status := h.serviceHandler.DeleteFleet(r.Context(), orgUUID, name)
	SetResponse(w, nil, status)
}

// (GET /api/v1/organizations/{orgID}/fleets/{name}/status)
func (h *TransportHandler) GetFleetStatus(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.GetFleetStatus(r.Context(), orgUUID, name)
	SetResponse(w, body, status)
}

// (PUT /api/v1/organizations/{orgID}/fleets/{name}/status)
func (h *TransportHandler) ReplaceFleetStatus(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	var fleet api.Fleet
	if err := json.NewDecoder(r.Body).Decode(&fleet); err != nil {
		SetParseFailureResponse(w, err)
		return
	}

	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.ReplaceFleetStatus(r.Context(), orgUUID, name, fleet)
	SetResponse(w, body, status)
}

// (PATCH /api/v1/organizations/{orgID}/fleets/{name})
func (h *TransportHandler) PatchFleet(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
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

	body, status := h.serviceHandler.PatchFleet(r.Context(), orgUUID, name, patch)
	SetResponse(w, body, status)
}

// (PATCH /api/v1/organizations/{orgID}/fleets/{name}/status)
func (h *TransportHandler) PatchFleetStatus(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	status := api.StatusNotImplemented("not yet implemented")
	SetResponse(w, nil, status)
}
