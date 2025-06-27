package authz

import (
	"context"
	"errors"
	"testing"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/flightctl/flightctl/internal/flterrors"
	"github.com/flightctl/flightctl/internal/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -source=basic_org_auth.go -destination=mock_organization_getter.go -package=authz

func TestBasicOrgAuth_CheckPermission(t *testing.T) {
	testOrgID := uuid.New()
	testOrg := &api.Organization{
		Id: &testOrgID,
	}

	tests := []struct {
		name           string
		setupContext   func() context.Context
		setupOrgGetter func(ctrl *gomock.Controller) OrganizationGetter
		expectedResult bool
		expectedError  error
	}{
		{
			name: "no organization ID in context",
			setupContext: func() context.Context {
				// Return a context without organization ID
				return context.Background()
			},
			setupOrgGetter: func(ctrl *gomock.Controller) OrganizationGetter {
				// Don't need to set up mock expectations since it shouldn't be called
				return NewMockOrganizationGetter(ctrl)
			},
			expectedResult: false,
			expectedError:  flterrors.ErrInvalidOrganizationID,
		},
		{
			name: "orgGetter is nil",
			setupContext: func() context.Context {
				return util.WithOrganizationID(context.Background(), testOrgID)
			},
			setupOrgGetter: func(ctrl *gomock.Controller) OrganizationGetter {
				return nil
			},
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name: "organization not found",
			setupContext: func() context.Context {
				return util.WithOrganizationID(context.Background(), testOrgID)
			},
			setupOrgGetter: func(ctrl *gomock.Controller) OrganizationGetter {
				mock := NewMockOrganizationGetter(ctrl)
				mock.EXPECT().Get(gomock.Any(), testOrgID).Return(nil, flterrors.ErrResourceNotFound)
				return mock
			},
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name: "error retrieving organization",
			setupContext: func() context.Context {
				return util.WithOrganizationID(context.Background(), testOrgID)
			},
			setupOrgGetter: func(ctrl *gomock.Controller) OrganizationGetter {
				mock := NewMockOrganizationGetter(ctrl)
				expectedErr := errors.New("database connection failed")
				mock.EXPECT().Get(gomock.Any(), testOrgID).Return(nil, expectedErr)
				return mock
			},
			expectedResult: false,
			expectedError:  errors.New("database connection failed"),
		},
		{
			name: "successful permission check",
			setupContext: func() context.Context {
				return util.WithOrganizationID(context.Background(), testOrgID)
			},
			setupOrgGetter: func(ctrl *gomock.Controller) OrganizationGetter {
				mock := NewMockOrganizationGetter(ctrl)
				mock.EXPECT().Get(gomock.Any(), testOrgID).Return(testOrg, nil)
				return mock
			},
			expectedResult: true,
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := tt.setupContext()
			orgGetter := tt.setupOrgGetter(ctrl)

			auth := NewBasicOrgAuth(orgGetter)

			result, err := auth.CheckPermission(ctx, "devices", "list")

			require.Equal(tt.expectedResult, result)
			if tt.expectedError != nil {
				require.Error(err)
				require.Equal(tt.expectedError.Error(), err.Error())
			} else {
				require.NoError(err)
			}
		})
	}
}
