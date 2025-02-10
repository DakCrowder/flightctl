// Code generated by MockGen. DO NOT EDIT.
// Source: spec.go
//
// Generated by this command:
//
//	mockgen -source=spec.go -destination=mock_spec.go -package=spec
//

// Package spec is a generated GoMock package.
package spec

import (
	context "context"
	reflect "reflect"

	v1alpha1 "github.com/flightctl/flightctl/api/v1alpha1"
	client "github.com/flightctl/flightctl/internal/agent/client"
	policy "github.com/flightctl/flightctl/internal/agent/device/policy"
	gomock "go.uber.org/mock/gomock"
)

// MockManager is a mock of Manager interface.
type MockManager struct {
	ctrl     *gomock.Controller
	recorder *MockManagerMockRecorder
}

// MockManagerMockRecorder is the mock recorder for MockManager.
type MockManagerMockRecorder struct {
	mock *MockManager
}

// NewMockManager creates a new mock instance.
func NewMockManager(ctrl *gomock.Controller) *MockManager {
	mock := &MockManager{ctrl: ctrl}
	mock.recorder = &MockManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockManager) EXPECT() *MockManagerMockRecorder {
	return m.recorder
}

// CheckOsReconciliation mocks base method.
func (m *MockManager) CheckOsReconciliation(ctx context.Context) (string, bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckOsReconciliation", ctx)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CheckOsReconciliation indicates an expected call of CheckOsReconciliation.
func (mr *MockManagerMockRecorder) CheckOsReconciliation(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckOsReconciliation", reflect.TypeOf((*MockManager)(nil).CheckOsReconciliation), ctx)
}

// CheckPolicy mocks base method.
func (m *MockManager) CheckPolicy(ctx context.Context, policyType policy.Type, version string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckPolicy", ctx, policyType, version)
	ret0, _ := ret[0].(error)
	return ret0
}

// CheckPolicy indicates an expected call of CheckPolicy.
func (mr *MockManagerMockRecorder) CheckPolicy(ctx, policyType, version any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckPolicy", reflect.TypeOf((*MockManager)(nil).CheckPolicy), ctx, policyType, version)
}

// ClearRollback mocks base method.
func (m *MockManager) ClearRollback() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ClearRollback")
	ret0, _ := ret[0].(error)
	return ret0
}

// ClearRollback indicates an expected call of ClearRollback.
func (mr *MockManagerMockRecorder) ClearRollback() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ClearRollback", reflect.TypeOf((*MockManager)(nil).ClearRollback))
}

// CreateRollback mocks base method.
func (m *MockManager) CreateRollback(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateRollback", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateRollback indicates an expected call of CreateRollback.
func (mr *MockManagerMockRecorder) CreateRollback(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateRollback", reflect.TypeOf((*MockManager)(nil).CreateRollback), ctx)
}

// Ensure mocks base method.
func (m *MockManager) Ensure() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ensure")
	ret0, _ := ret[0].(error)
	return ret0
}

// Ensure indicates an expected call of Ensure.
func (mr *MockManagerMockRecorder) Ensure() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ensure", reflect.TypeOf((*MockManager)(nil).Ensure))
}

// GetDesired mocks base method.
func (m *MockManager) GetDesired(ctx context.Context) (*v1alpha1.RenderedDeviceSpec, bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDesired", ctx)
	ret0, _ := ret[0].(*v1alpha1.RenderedDeviceSpec)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetDesired indicates an expected call of GetDesired.
func (mr *MockManagerMockRecorder) GetDesired(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDesired", reflect.TypeOf((*MockManager)(nil).GetDesired), ctx)
}

// Initialize mocks base method.
func (m *MockManager) Initialize(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Initialize", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Initialize indicates an expected call of Initialize.
func (mr *MockManagerMockRecorder) Initialize(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Initialize", reflect.TypeOf((*MockManager)(nil).Initialize), ctx)
}

// IsOSUpdate mocks base method.
func (m *MockManager) IsOSUpdate() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsOSUpdate")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsOSUpdate indicates an expected call of IsOSUpdate.
func (mr *MockManagerMockRecorder) IsOSUpdate() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsOSUpdate", reflect.TypeOf((*MockManager)(nil).IsOSUpdate))
}

// IsRollingBack mocks base method.
func (m *MockManager) IsRollingBack(ctx context.Context) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsRollingBack", ctx)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsRollingBack indicates an expected call of IsRollingBack.
func (mr *MockManagerMockRecorder) IsRollingBack(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsRollingBack", reflect.TypeOf((*MockManager)(nil).IsRollingBack), ctx)
}

// IsUpgrading mocks base method.
func (m *MockManager) IsUpgrading() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsUpgrading")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsUpgrading indicates an expected call of IsUpgrading.
func (mr *MockManagerMockRecorder) IsUpgrading() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsUpgrading", reflect.TypeOf((*MockManager)(nil).IsUpgrading))
}

// OSVersion mocks base method.
func (m *MockManager) OSVersion(specType Type) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OSVersion", specType)
	ret0, _ := ret[0].(string)
	return ret0
}

// OSVersion indicates an expected call of OSVersion.
func (mr *MockManagerMockRecorder) OSVersion(specType any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OSVersion", reflect.TypeOf((*MockManager)(nil).OSVersion), specType)
}

// Read mocks base method.
func (m *MockManager) Read(specType Type) (*v1alpha1.RenderedDeviceSpec, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", specType)
	ret0, _ := ret[0].(*v1alpha1.RenderedDeviceSpec)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read.
func (mr *MockManagerMockRecorder) Read(specType any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockManager)(nil).Read), specType)
}

// RenderedVersion mocks base method.
func (m *MockManager) RenderedVersion(specType Type) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RenderedVersion", specType)
	ret0, _ := ret[0].(string)
	return ret0
}

// RenderedVersion indicates an expected call of RenderedVersion.
func (mr *MockManagerMockRecorder) RenderedVersion(specType any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RenderedVersion", reflect.TypeOf((*MockManager)(nil).RenderedVersion), specType)
}

// Rollback mocks base method.
func (m *MockManager) Rollback(ctx context.Context, opts ...RollbackOption) error {
	m.ctrl.T.Helper()
	varargs := []any{ctx}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Rollback", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Rollback indicates an expected call of Rollback.
func (mr *MockManagerMockRecorder) Rollback(ctx any, opts ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Rollback", reflect.TypeOf((*MockManager)(nil).Rollback), varargs...)
}

// SetClient mocks base method.
func (m *MockManager) SetClient(arg0 client.Management) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetClient", arg0)
}

// SetClient indicates an expected call of SetClient.
func (mr *MockManagerMockRecorder) SetClient(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetClient", reflect.TypeOf((*MockManager)(nil).SetClient), arg0)
}

// SetUpgradeFailed mocks base method.
func (m *MockManager) SetUpgradeFailed(version string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetUpgradeFailed", version)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetUpgradeFailed indicates an expected call of SetUpgradeFailed.
func (mr *MockManagerMockRecorder) SetUpgradeFailed(version any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetUpgradeFailed", reflect.TypeOf((*MockManager)(nil).SetUpgradeFailed), version)
}

// Status mocks base method.
func (m *MockManager) Status(arg0 context.Context, arg1 *v1alpha1.DeviceStatus) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Status", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Status indicates an expected call of Status.
func (mr *MockManagerMockRecorder) Status(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Status", reflect.TypeOf((*MockManager)(nil).Status), arg0, arg1)
}

// Upgrade mocks base method.
func (m *MockManager) Upgrade(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Upgrade", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Upgrade indicates an expected call of Upgrade.
func (mr *MockManagerMockRecorder) Upgrade(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Upgrade", reflect.TypeOf((*MockManager)(nil).Upgrade), ctx)
}

// MockPriorityQueue is a mock of PriorityQueue interface.
type MockPriorityQueue struct {
	ctrl     *gomock.Controller
	recorder *MockPriorityQueueMockRecorder
}

// MockPriorityQueueMockRecorder is the mock recorder for MockPriorityQueue.
type MockPriorityQueueMockRecorder struct {
	mock *MockPriorityQueue
}

// NewMockPriorityQueue creates a new mock instance.
func NewMockPriorityQueue(ctrl *gomock.Controller) *MockPriorityQueue {
	mock := &MockPriorityQueue{ctrl: ctrl}
	mock.recorder = &MockPriorityQueueMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPriorityQueue) EXPECT() *MockPriorityQueueMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockPriorityQueue) Add(ctx context.Context, spec *v1alpha1.RenderedDeviceSpec) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Add", ctx, spec)
}

// Add indicates an expected call of Add.
func (mr *MockPriorityQueueMockRecorder) Add(ctx, spec any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockPriorityQueue)(nil).Add), ctx, spec)
}

// CheckPolicy mocks base method.
func (m *MockPriorityQueue) CheckPolicy(ctx context.Context, policyType policy.Type, version string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckPolicy", ctx, policyType, version)
	ret0, _ := ret[0].(error)
	return ret0
}

// CheckPolicy indicates an expected call of CheckPolicy.
func (mr *MockPriorityQueueMockRecorder) CheckPolicy(ctx, policyType, version any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckPolicy", reflect.TypeOf((*MockPriorityQueue)(nil).CheckPolicy), ctx, policyType, version)
}

// IsFailed mocks base method.
func (m *MockPriorityQueue) IsFailed(version int64) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsFailed", version)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsFailed indicates an expected call of IsFailed.
func (mr *MockPriorityQueueMockRecorder) IsFailed(version any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsFailed", reflect.TypeOf((*MockPriorityQueue)(nil).IsFailed), version)
}

// Next mocks base method.
func (m *MockPriorityQueue) Next(ctx context.Context) (*v1alpha1.RenderedDeviceSpec, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Next", ctx)
	ret0, _ := ret[0].(*v1alpha1.RenderedDeviceSpec)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// Next indicates an expected call of Next.
func (mr *MockPriorityQueueMockRecorder) Next(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Next", reflect.TypeOf((*MockPriorityQueue)(nil).Next), ctx)
}

// Remove mocks base method.
func (m *MockPriorityQueue) Remove(version int64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Remove", version)
}

// Remove indicates an expected call of Remove.
func (mr *MockPriorityQueueMockRecorder) Remove(version any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Remove", reflect.TypeOf((*MockPriorityQueue)(nil).Remove), version)
}

// SetFailed mocks base method.
func (m *MockPriorityQueue) SetFailed(version int64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetFailed", version)
}

// SetFailed indicates an expected call of SetFailed.
func (mr *MockPriorityQueueMockRecorder) SetFailed(version any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetFailed", reflect.TypeOf((*MockPriorityQueue)(nil).SetFailed), version)
}
