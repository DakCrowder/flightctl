package auth

import (
	"context"
	"testing"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/flightctl/flightctl/internal/auth/common"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestOrganizationValidationSkip(t *testing.T) {
	require := require.New(t)

	// Mock organization getter that would fail if called
	mockOrgGetter := &failingOrgGetter{}
	validator := common.NewOrganizationExistsValidator(mockOrgGetter)
	
	// Test case 1: organizations list should skip validation
	err := validator.ValidateOrganization(context.Background(), "organizations", "list")
	require.NoError(err, "organizations list should skip organization validation")

	// Test case 2: other resources should attempt validation (and fail with our mock)
	err = validator.ValidateOrganization(context.Background(), "devices", "list")
	require.Error(err, "other resources should attempt organization validation")
	require.Contains(err.Error(), "no org id in context")

	// Test case 3: organizations with non-list action should attempt validation
	err = validator.ValidateOrganization(context.Background(), "organizations", "get")
	require.Error(err, "organizations get should attempt organization validation")
	require.Contains(err.Error(), "no org id in context")
}

type failingOrgGetter struct{}

func (f *failingOrgGetter) GetOrganization(ctx context.Context, orgID uuid.UUID) (*api.Organization, api.Status) {
	panic("should not be called in skip test")
}