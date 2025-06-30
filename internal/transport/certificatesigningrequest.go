package transport

import (
	"encoding/json"
	"net/http"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/oapi-codegen/runtime/types"
)

// (GET /api/v1/organizations/{orgID}/certificatesigningrequests)
func (h *TransportHandler) ListCertificateSigningRequests(w http.ResponseWriter, r *http.Request, orgID types.UUID, params api.ListCertificateSigningRequestsParams) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.ListCertificateSigningRequests(r.Context(), orgUUID, params)
	SetResponse(w, body, status)
}

// (POST /api/v1/organizations/{orgID}/certificatesigningrequests)
func (h *TransportHandler) CreateCertificateSigningRequest(w http.ResponseWriter, r *http.Request, orgID types.UUID) {
	var csr api.CertificateSigningRequest
	if err := json.NewDecoder(r.Body).Decode(&csr); err != nil {
		SetParseFailureResponse(w, err)
		return
	}

	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.CreateCertificateSigningRequest(r.Context(), orgUUID, csr)
	SetResponse(w, body, status)
}

// (DELETE /api/v1/organizations/{orgID}/certificatesigningrequests/{name})
func (h *TransportHandler) DeleteCertificateSigningRequest(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	status := h.serviceHandler.DeleteCertificateSigningRequest(r.Context(), orgUUID, name)
	SetResponse(w, nil, status)
}

// (GET /api/v1/organizations/{orgID}/certificatesigningrequests/{name})
func (h *TransportHandler) GetCertificateSigningRequest(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.GetCertificateSigningRequest(r.Context(), orgUUID, name)
	SetResponse(w, body, status)
}

// (PATCH /api/v1/organizations/{orgID}/certificatesigningrequests/{name})
func (h *TransportHandler) PatchCertificateSigningRequest(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
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

	body, status := h.serviceHandler.PatchCertificateSigningRequest(r.Context(), orgUUID, name, patch)
	SetResponse(w, body, status)
}

// (PUT /api/v1/organizations/{orgID}/certificatesigningrequests/{name})
func (h *TransportHandler) ReplaceCertificateSigningRequest(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	var csr api.CertificateSigningRequest
	if err := json.NewDecoder(r.Body).Decode(&csr); err != nil {
		SetParseFailureResponse(w, err)
		return
	}

	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.ReplaceCertificateSigningRequest(r.Context(), orgUUID, name, csr)
	SetResponse(w, body, status)
}

// (PUT /api/v1/organizations/{orgID}/certificatesigningrequests/{name}/approval)
func (h *TransportHandler) UpdateCertificateSigningRequestApproval(w http.ResponseWriter, r *http.Request, orgID types.UUID, name string) {
	var csr api.CertificateSigningRequest
	if err := json.NewDecoder(r.Body).Decode(&csr); err != nil {
		SetParseFailureResponse(w, err)
		return
	}

	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.UpdateCertificateSigningRequestApproval(r.Context(), orgUUID, name, csr)
	SetResponse(w, body, status)
}
