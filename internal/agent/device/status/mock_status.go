// Code generated by MockGen. DO NOT EDIT.
// Source: internal/agent/device/status/status.go
//
// Generated by this command:
//
//	mockgen -source=internal/agent/device/status/status.go -destination=internal/agent/device/status/mock_status.go -package=status
//

// Package status is a generated GoMock package.
package status

import (
	context "context"
	reflect "reflect"

	v1alpha1 "github.com/flightctl/flightctl/api/v1alpha1"
	client "github.com/flightctl/flightctl/internal/agent/client"
	gomock "go.uber.org/mock/gomock"
)

// MockExporter is a mock of Exporter interface.
type MockExporter struct {
	ctrl     *gomock.Controller
	recorder *MockExporterMockRecorder
}

// MockExporterMockRecorder is the mock recorder for MockExporter.
type MockExporterMockRecorder struct {
	mock *MockExporter
}

// NewMockExporter creates a new mock instance.
func NewMockExporter(ctrl *gomock.Controller) *MockExporter {
	mock := &MockExporter{ctrl: ctrl}
	mock.recorder = &MockExporterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExporter) EXPECT() *MockExporterMockRecorder {
	return m.recorder
}

// Export mocks base method.
func (m *MockExporter) Export(ctx context.Context, device *v1alpha1.DeviceStatus) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Export", ctx, device)
	ret0, _ := ret[0].(error)
	return ret0
}

// Export indicates an expected call of Export.
func (mr *MockExporterMockRecorder) Export(ctx, device any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Export", reflect.TypeOf((*MockExporter)(nil).Export), ctx, device)
}

// SetProperties mocks base method.
func (m *MockExporter) SetProperties(arg0 *v1alpha1.RenderedDeviceSpec) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetProperties", arg0)
}

// SetProperties indicates an expected call of SetProperties.
func (mr *MockExporterMockRecorder) SetProperties(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetProperties", reflect.TypeOf((*MockExporter)(nil).SetProperties), arg0)
}

// MockCollector is a mock of Collector interface.
type MockCollector struct {
	ctrl     *gomock.Controller
	recorder *MockCollectorMockRecorder
}

// MockCollectorMockRecorder is the mock recorder for MockCollector.
type MockCollectorMockRecorder struct {
	mock *MockCollector
}

// NewMockCollector creates a new mock instance.
func NewMockCollector(ctrl *gomock.Controller) *MockCollector {
	mock := &MockCollector{ctrl: ctrl}
	mock.recorder = &MockCollectorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCollector) EXPECT() *MockCollectorMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockCollector) Get(arg0 context.Context) *v1alpha1.DeviceStatus {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0)
	ret0, _ := ret[0].(*v1alpha1.DeviceStatus)
	return ret0
}

// Get indicates an expected call of Get.
func (mr *MockCollectorMockRecorder) Get(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockCollector)(nil).Get), arg0)
}

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

// Get mocks base method.
func (m *MockManager) Get(arg0 context.Context) *v1alpha1.DeviceStatus {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0)
	ret0, _ := ret[0].(*v1alpha1.DeviceStatus)
	return ret0
}

// Get indicates an expected call of Get.
func (mr *MockManagerMockRecorder) Get(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockManager)(nil).Get), arg0)
}

// SetClient mocks base method.
func (m *MockManager) SetClient(arg0 *client.Management) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetClient", arg0)
}

// SetClient indicates an expected call of SetClient.
func (mr *MockManagerMockRecorder) SetClient(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetClient", reflect.TypeOf((*MockManager)(nil).SetClient), arg0)
}

// SetProperties mocks base method.
func (m *MockManager) SetProperties(arg0 *v1alpha1.RenderedDeviceSpec) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetProperties", arg0)
}

// SetProperties indicates an expected call of SetProperties.
func (mr *MockManagerMockRecorder) SetProperties(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetProperties", reflect.TypeOf((*MockManager)(nil).SetProperties), arg0)
}

// Sync mocks base method.
func (m *MockManager) Sync(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Sync", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Sync indicates an expected call of Sync.
func (mr *MockManagerMockRecorder) Sync(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Sync", reflect.TypeOf((*MockManager)(nil).Sync), arg0)
}

// Update mocks base method.
func (m *MockManager) Update(ctx context.Context, updateFuncs ...UpdateStatusFn) (*v1alpha1.DeviceStatus, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx}
	for _, a := range updateFuncs {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Update", varargs...)
	ret0, _ := ret[0].(*v1alpha1.DeviceStatus)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Update indicates an expected call of Update.
func (mr *MockManagerMockRecorder) Update(ctx any, updateFuncs ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx}, updateFuncs...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockManager)(nil).Update), varargs...)
}

// UpdateCondition mocks base method.
func (m *MockManager) UpdateCondition(arg0 context.Context, arg1 v1alpha1.Condition) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateCondition", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateCondition indicates an expected call of UpdateCondition.
func (mr *MockManagerMockRecorder) UpdateCondition(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateCondition", reflect.TypeOf((*MockManager)(nil).UpdateCondition), arg0, arg1)
}
