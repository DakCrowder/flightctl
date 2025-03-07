// Code generated by MockGen. DO NOT EDIT.
// Source: router_grpc.pb.go
//
// Generated by this command:
//
//	mockgen -source=router_grpc.pb.go -destination=../../../internal/agent/device/console/mock_router_service_client.go -package=console
//

// Package console is a generated GoMock package.
package console

import (
	context "context"
	reflect "reflect"

	grpc_v1 "github.com/flightctl/flightctl/api/grpc/v1"
	gomock "go.uber.org/mock/gomock"
	grpc "google.golang.org/grpc"
	metadata "google.golang.org/grpc/metadata"
)

// MockRouterServiceClient is a mock of RouterServiceClient interface.
type MockRouterServiceClient struct {
	ctrl     *gomock.Controller
	recorder *MockRouterServiceClientMockRecorder
}

// MockRouterServiceClientMockRecorder is the mock recorder for MockRouterServiceClient.
type MockRouterServiceClientMockRecorder struct {
	mock *MockRouterServiceClient
}

// NewMockRouterServiceClient creates a new mock instance.
func NewMockRouterServiceClient(ctrl *gomock.Controller) *MockRouterServiceClient {
	mock := &MockRouterServiceClient{ctrl: ctrl}
	mock.recorder = &MockRouterServiceClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRouterServiceClient) EXPECT() *MockRouterServiceClientMockRecorder {
	return m.recorder
}

// Stream mocks base method.
func (m *MockRouterServiceClient) Stream(ctx context.Context, opts ...grpc.CallOption) (grpc_v1.RouterService_StreamClient, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Stream", varargs...)
	ret0, _ := ret[0].(grpc_v1.RouterService_StreamClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Stream indicates an expected call of Stream.
func (mr *MockRouterServiceClientMockRecorder) Stream(ctx any, opts ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stream", reflect.TypeOf((*MockRouterServiceClient)(nil).Stream), varargs...)
}

// MockRouterService_StreamClient is a mock of RouterService_StreamClient interface.
type MockRouterService_StreamClient struct {
	ctrl     *gomock.Controller
	recorder *MockRouterService_StreamClientMockRecorder
}

// MockRouterService_StreamClientMockRecorder is the mock recorder for MockRouterService_StreamClient.
type MockRouterService_StreamClientMockRecorder struct {
	mock *MockRouterService_StreamClient
}

// NewMockRouterService_StreamClient creates a new mock instance.
func NewMockRouterService_StreamClient(ctrl *gomock.Controller) *MockRouterService_StreamClient {
	mock := &MockRouterService_StreamClient{ctrl: ctrl}
	mock.recorder = &MockRouterService_StreamClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRouterService_StreamClient) EXPECT() *MockRouterService_StreamClientMockRecorder {
	return m.recorder
}

// CloseSend mocks base method.
func (m *MockRouterService_StreamClient) CloseSend() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CloseSend")
	ret0, _ := ret[0].(error)
	return ret0
}

// CloseSend indicates an expected call of CloseSend.
func (mr *MockRouterService_StreamClientMockRecorder) CloseSend() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CloseSend", reflect.TypeOf((*MockRouterService_StreamClient)(nil).CloseSend))
}

// Context mocks base method.
func (m *MockRouterService_StreamClient) Context() context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context.
func (mr *MockRouterService_StreamClientMockRecorder) Context() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockRouterService_StreamClient)(nil).Context))
}

// Header mocks base method.
func (m *MockRouterService_StreamClient) Header() (metadata.MD, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Header")
	ret0, _ := ret[0].(metadata.MD)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Header indicates an expected call of Header.
func (mr *MockRouterService_StreamClientMockRecorder) Header() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Header", reflect.TypeOf((*MockRouterService_StreamClient)(nil).Header))
}

// Recv mocks base method.
func (m *MockRouterService_StreamClient) Recv() (*grpc_v1.StreamResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Recv")
	ret0, _ := ret[0].(*grpc_v1.StreamResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Recv indicates an expected call of Recv.
func (mr *MockRouterService_StreamClientMockRecorder) Recv() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Recv", reflect.TypeOf((*MockRouterService_StreamClient)(nil).Recv))
}

// RecvMsg mocks base method.
func (m_2 *MockRouterService_StreamClient) RecvMsg(m any) error {
	m_2.ctrl.T.Helper()
	ret := m_2.ctrl.Call(m_2, "RecvMsg", m)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg.
func (mr *MockRouterService_StreamClientMockRecorder) RecvMsg(m any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockRouterService_StreamClient)(nil).RecvMsg), m)
}

// Send mocks base method.
func (m *MockRouterService_StreamClient) Send(arg0 *grpc_v1.StreamRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MockRouterService_StreamClientMockRecorder) Send(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockRouterService_StreamClient)(nil).Send), arg0)
}

// SendMsg mocks base method.
func (m_2 *MockRouterService_StreamClient) SendMsg(m any) error {
	m_2.ctrl.T.Helper()
	ret := m_2.ctrl.Call(m_2, "SendMsg", m)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg.
func (mr *MockRouterService_StreamClientMockRecorder) SendMsg(m any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockRouterService_StreamClient)(nil).SendMsg), m)
}

// Trailer mocks base method.
func (m *MockRouterService_StreamClient) Trailer() metadata.MD {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Trailer")
	ret0, _ := ret[0].(metadata.MD)
	return ret0
}

// Trailer indicates an expected call of Trailer.
func (mr *MockRouterService_StreamClientMockRecorder) Trailer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Trailer", reflect.TypeOf((*MockRouterService_StreamClient)(nil).Trailer))
}

// MockRouterServiceServer is a mock of RouterServiceServer interface.
type MockRouterServiceServer struct {
	ctrl     *gomock.Controller
	recorder *MockRouterServiceServerMockRecorder
}

// MockRouterServiceServerMockRecorder is the mock recorder for MockRouterServiceServer.
type MockRouterServiceServerMockRecorder struct {
	mock *MockRouterServiceServer
}

// NewMockRouterServiceServer creates a new mock instance.
func NewMockRouterServiceServer(ctrl *gomock.Controller) *MockRouterServiceServer {
	mock := &MockRouterServiceServer{ctrl: ctrl}
	mock.recorder = &MockRouterServiceServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRouterServiceServer) EXPECT() *MockRouterServiceServerMockRecorder {
	return m.recorder
}

// Stream mocks base method.
func (m *MockRouterServiceServer) Stream(arg0 grpc_v1.RouterService_StreamServer) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stream", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Stream indicates an expected call of Stream.
func (mr *MockRouterServiceServerMockRecorder) Stream(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stream", reflect.TypeOf((*MockRouterServiceServer)(nil).Stream), arg0)
}

// mustEmbedUnimplementedRouterServiceServer mocks base method.
func (m *MockRouterServiceServer) mustEmbedUnimplementedRouterServiceServer() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "mustEmbedUnimplementedRouterServiceServer")
}

// mustEmbedUnimplementedRouterServiceServer indicates an expected call of mustEmbedUnimplementedRouterServiceServer.
func (mr *MockRouterServiceServerMockRecorder) mustEmbedUnimplementedRouterServiceServer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "mustEmbedUnimplementedRouterServiceServer", reflect.TypeOf((*MockRouterServiceServer)(nil).mustEmbedUnimplementedRouterServiceServer))
}

// MockUnsafeRouterServiceServer is a mock of UnsafeRouterServiceServer interface.
type MockUnsafeRouterServiceServer struct {
	ctrl     *gomock.Controller
	recorder *MockUnsafeRouterServiceServerMockRecorder
}

// MockUnsafeRouterServiceServerMockRecorder is the mock recorder for MockUnsafeRouterServiceServer.
type MockUnsafeRouterServiceServerMockRecorder struct {
	mock *MockUnsafeRouterServiceServer
}

// NewMockUnsafeRouterServiceServer creates a new mock instance.
func NewMockUnsafeRouterServiceServer(ctrl *gomock.Controller) *MockUnsafeRouterServiceServer {
	mock := &MockUnsafeRouterServiceServer{ctrl: ctrl}
	mock.recorder = &MockUnsafeRouterServiceServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUnsafeRouterServiceServer) EXPECT() *MockUnsafeRouterServiceServerMockRecorder {
	return m.recorder
}

// mustEmbedUnimplementedRouterServiceServer mocks base method.
func (m *MockUnsafeRouterServiceServer) mustEmbedUnimplementedRouterServiceServer() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "mustEmbedUnimplementedRouterServiceServer")
}

// mustEmbedUnimplementedRouterServiceServer indicates an expected call of mustEmbedUnimplementedRouterServiceServer.
func (mr *MockUnsafeRouterServiceServerMockRecorder) mustEmbedUnimplementedRouterServiceServer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "mustEmbedUnimplementedRouterServiceServer", reflect.TypeOf((*MockUnsafeRouterServiceServer)(nil).mustEmbedUnimplementedRouterServiceServer))
}

// MockRouterService_StreamServer is a mock of RouterService_StreamServer interface.
type MockRouterService_StreamServer struct {
	ctrl     *gomock.Controller
	recorder *MockRouterService_StreamServerMockRecorder
}

// MockRouterService_StreamServerMockRecorder is the mock recorder for MockRouterService_StreamServer.
type MockRouterService_StreamServerMockRecorder struct {
	mock *MockRouterService_StreamServer
}

// NewMockRouterService_StreamServer creates a new mock instance.
func NewMockRouterService_StreamServer(ctrl *gomock.Controller) *MockRouterService_StreamServer {
	mock := &MockRouterService_StreamServer{ctrl: ctrl}
	mock.recorder = &MockRouterService_StreamServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRouterService_StreamServer) EXPECT() *MockRouterService_StreamServerMockRecorder {
	return m.recorder
}

// Context mocks base method.
func (m *MockRouterService_StreamServer) Context() context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context.
func (mr *MockRouterService_StreamServerMockRecorder) Context() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockRouterService_StreamServer)(nil).Context))
}

// Recv mocks base method.
func (m *MockRouterService_StreamServer) Recv() (*grpc_v1.StreamRequest, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Recv")
	ret0, _ := ret[0].(*grpc_v1.StreamRequest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Recv indicates an expected call of Recv.
func (mr *MockRouterService_StreamServerMockRecorder) Recv() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Recv", reflect.TypeOf((*MockRouterService_StreamServer)(nil).Recv))
}

// RecvMsg mocks base method.
func (m_2 *MockRouterService_StreamServer) RecvMsg(m any) error {
	m_2.ctrl.T.Helper()
	ret := m_2.ctrl.Call(m_2, "RecvMsg", m)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg.
func (mr *MockRouterService_StreamServerMockRecorder) RecvMsg(m any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockRouterService_StreamServer)(nil).RecvMsg), m)
}

// Send mocks base method.
func (m *MockRouterService_StreamServer) Send(arg0 *grpc_v1.StreamResponse) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MockRouterService_StreamServerMockRecorder) Send(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockRouterService_StreamServer)(nil).Send), arg0)
}

// SendHeader mocks base method.
func (m *MockRouterService_StreamServer) SendHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendHeader indicates an expected call of SendHeader.
func (mr *MockRouterService_StreamServerMockRecorder) SendHeader(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHeader", reflect.TypeOf((*MockRouterService_StreamServer)(nil).SendHeader), arg0)
}

// SendMsg mocks base method.
func (m_2 *MockRouterService_StreamServer) SendMsg(m any) error {
	m_2.ctrl.T.Helper()
	ret := m_2.ctrl.Call(m_2, "SendMsg", m)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg.
func (mr *MockRouterService_StreamServerMockRecorder) SendMsg(m any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockRouterService_StreamServer)(nil).SendMsg), m)
}

// SetHeader mocks base method.
func (m *MockRouterService_StreamServer) SetHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeader indicates an expected call of SetHeader.
func (mr *MockRouterService_StreamServerMockRecorder) SetHeader(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeader", reflect.TypeOf((*MockRouterService_StreamServer)(nil).SetHeader), arg0)
}

// SetTrailer mocks base method.
func (m *MockRouterService_StreamServer) SetTrailer(arg0 metadata.MD) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetTrailer", arg0)
}

// SetTrailer indicates an expected call of SetTrailer.
func (mr *MockRouterService_StreamServerMockRecorder) SetTrailer(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTrailer", reflect.TypeOf((*MockRouterService_StreamServer)(nil).SetTrailer), arg0)
}
