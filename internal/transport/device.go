package transport

import (
	"encoding/json"
	"net/http"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/oapi-codegen/runtime/types"
)

// (POST /api/v1/organizations/{orgID}/devices)
func (h *TransportHandler) CreateDevice(w http.ResponseWriter, r *http.Request, orgID types.UUID) {
	var device api.Device
	if err := json.NewDecoder(r.Body).Decode(&device); err != nil {
		SetParseFailureResponse(w, err)
		return
	}

	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.CreateDevice(r.Context(), orgUUID, device)
	SetResponse(w, body, status)
}

// (GET /api/v1/organizations/{orgID}/devices)
func (h *TransportHandler) ListDevices(w http.ResponseWriter, r *http.Request, orgID types.UUID, params api.ListDevicesParams) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.ListDevices(r.Context(), orgUUID, params, nil)
	SetResponse(w, body, status)
}

// (GET /api/v1/organizations/{orgID}/devices/{name})
func (h *TransportHandler) GetDevice(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.GetDevice(r.Context(), orgUUID, name)
	SetResponse(w, body, status)
}

// (PUT /api/v1/organizations/{orgID}/devices/{name})
func (h *TransportHandler) ReplaceDevice(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	var device api.Device
	if err := json.NewDecoder(r.Body).Decode(&device); err != nil {
		SetParseFailureResponse(w, err)
		return
	}

	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.ReplaceDevice(r.Context(), orgUUID, name, device, nil)
	SetResponse(w, body, status)
}

// (DELETE /api/v1/organizations/{orgID}/devices/{name})
func (h *TransportHandler) DeleteDevice(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	status := h.serviceHandler.DeleteDevice(r.Context(), orgUUID, name)
	SetResponse(w, nil, status)
}

// (GET /api/v1/organizations/{orgID}/devices/{name}/status)
func (h *TransportHandler) GetDeviceStatus(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.GetDeviceStatus(r.Context(), orgUUID, name)
	SetResponse(w, body, status)
}

// (PUT /api/v1/organizations/{orgID}/devices/{name}/status)
func (h *TransportHandler) ReplaceDeviceStatus(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	var device api.Device
	if err := json.NewDecoder(r.Body).Decode(&device); err != nil {
		SetParseFailureResponse(w, err)
		return
	}

	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.ReplaceDeviceStatus(r.Context(), orgUUID, name, device)
	SetResponse(w, body, status)
}

// (GET /api/v1/organizations/{orgID}/devices/{name}/rendered)
func (h *TransportHandler) GetRenderedDevice(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string, params api.GetRenderedDeviceParams) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.GetRenderedDevice(r.Context(), orgUUID, name, params)
	SetResponse(w, body, status)
}

// (PATCH /api/v1/organizations/{orgID}/devices/{name})
func (h *TransportHandler) PatchDevice(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
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

	body, status := h.serviceHandler.PatchDevice(r.Context(), orgUUID, name, patch)
	SetResponse(w, body, status)
}

// (PATCH /api/v1/organizations/{orgID}/devices/{name}/status)
func (h *TransportHandler) PatchDeviceStatus(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
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

	body, status := h.serviceHandler.PatchDeviceStatus(r.Context(), orgUUID, name, patch)
	SetResponse(w, body, status)
}

// (PUT /api/v1/organizations/{orgID}/devices/{name}/decommission)
func (h *TransportHandler) DecommissionDevice(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	var decom api.DeviceDecommission
	if err := json.NewDecoder(r.Body).Decode(&decom); err != nil {
		SetParseFailureResponse(w, err)
		return
	}

	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.DecommissionDevice(r.Context(), orgUUID, name, decom)
	SetResponse(w, body, status)
}
