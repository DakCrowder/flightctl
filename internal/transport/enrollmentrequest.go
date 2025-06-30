package transport

import (
	"encoding/json"
	"net/http"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/oapi-codegen/runtime/types"
)

// (POST /api/v1/organizations/{orgID}/enrollmentrequests)
func (h *TransportHandler) CreateEnrollmentRequest(w http.ResponseWriter, r *http.Request, orgID types.UUID) {
	var enrollmentRequest api.EnrollmentRequest
	if err := json.NewDecoder(r.Body).Decode(&enrollmentRequest); err != nil {
		SetParseFailureResponse(w, err)
		return
	}

	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.CreateEnrollmentRequest(r.Context(), orgUUID, enrollmentRequest)
	SetResponse(w, body, status)
}

// (GET /api/v1/organizations/{orgID}/enrollmentrequests)
func (h *TransportHandler) ListEnrollmentRequests(w http.ResponseWriter, r *http.Request, orgID types.UUID, params api.ListEnrollmentRequestsParams) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.ListEnrollmentRequests(r.Context(), orgUUID, params)
	SetResponse(w, body, status)
}

// (GET /api/v1/organizations/{orgID}/enrollmentrequests/{name})
func (h *TransportHandler) GetEnrollmentRequest(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.GetEnrollmentRequest(r.Context(), orgUUID, name)
	SetResponse(w, body, status)
}

// (PUT /api/v1/organizations/{orgID}/enrollmentrequests/{name})
func (h *TransportHandler) ReplaceEnrollmentRequest(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	var enrollmentRequest api.EnrollmentRequest
	if err := json.NewDecoder(r.Body).Decode(&enrollmentRequest); err != nil {
		SetParseFailureResponse(w, err)
		return
	}

	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.ReplaceEnrollmentRequest(r.Context(), orgUUID, name, enrollmentRequest)
	SetResponse(w, body, status)
}

// (DELETE /api/v1/organizations/{orgID}/enrollmentrequests/{name})
func (h *TransportHandler) DeleteEnrollmentRequest(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	status := h.serviceHandler.DeleteEnrollmentRequest(r.Context(), orgUUID, name)
	SetResponse(w, nil, status)
}

// (GET /api/v1/organizations/{orgID}/enrollmentrequests/{name}/status)
func (h *TransportHandler) GetEnrollmentRequestStatus(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.GetEnrollmentRequestStatus(r.Context(), orgUUID, name)
	SetResponse(w, body, status)
}

// (PUT /api/v1/organizations/{orgID}/enrollmentrequests/{name}/status)
func (h *TransportHandler) ReplaceEnrollmentRequestStatus(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	var enrollmentRequest api.EnrollmentRequest
	if err := json.NewDecoder(r.Body).Decode(&enrollmentRequest); err != nil {
		SetParseFailureResponse(w, err)
		return
	}

	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.ReplaceEnrollmentRequestStatus(r.Context(), orgUUID, name, enrollmentRequest)
	SetResponse(w, body, status)
}

// (PATCH /api/v1/organizations/{orgID}/enrollmentrequests/{name})
func (h *TransportHandler) PatchEnrollmentRequest(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
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

	body, status := h.serviceHandler.PatchEnrollmentRequest(r.Context(), orgUUID, name, patch)
	SetResponse(w, body, status)
}

// (PATCH /api/v1/organizations/{orgID}/enrollmentrequests/{name}/status)
func (h *TransportHandler) PatchEnrollmentRequestStatus(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	status := api.StatusNotImplemented("not yet implemented")
	SetResponse(w, nil, status)
}

// (PUT /api/v1/organizations/{orgID}/enrollmentrequests/{name}/approval)
func (h *TransportHandler) ApproveEnrollmentRequest(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	var approval api.EnrollmentRequestApproval
	if err := json.NewDecoder(r.Body).Decode(&approval); err != nil {
		SetParseFailureResponse(w, err)
		return
	}

	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.ApproveEnrollmentRequest(r.Context(), orgUUID, name, approval)
	SetResponse(w, body, status)
}
