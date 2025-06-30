package transport

import (
	"net/http"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/oapi-codegen/runtime/types"
)

// (GET /api/v1/organizations/{orgID}/labels)
func (h *TransportHandler) ListLabels(w http.ResponseWriter, r *http.Request, orgID types.UUID, params api.ListLabelsParams) {
	orgUUID, err := convertOrgID(orgID)
	if err != nil {
		SetResponse(w, nil, api.StatusBadRequest("invalid organization ID"))
		return
	}

	body, status := h.serviceHandler.ListLabels(r.Context(), orgUUID, params)
	SetResponse(w, body, status)
}
