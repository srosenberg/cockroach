// Package roachpbmock is a generated GoMock package.
package roachpbmock

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	context "context"
	reflect "reflect"

	roachpb "github.com/cockroachdb/cockroach/pkg/roachpb"
	gomock "github.com/golang/mock/gomock"
	grpc "google.golang.org/grpc"
	metadata "google.golang.org/grpc/metadata"
)

type MockInternalClient struct {
	ctrl     *gomock.Controller
	recorder *MockInternalClientMockRecorder
}

type MockInternalClientMockRecorder struct {
	mock *MockInternalClient
}

func NewMockInternalClient(ctrl *gomock.Controller) *MockInternalClient {
	__antithesis_instrumentation__.Notify(176882)
	mock := &MockInternalClient{ctrl: ctrl}
	mock.recorder = &MockInternalClientMockRecorder{mock}
	return mock
}

func (m *MockInternalClient) EXPECT() *MockInternalClientMockRecorder {
	__antithesis_instrumentation__.Notify(176883)
	return m.recorder
}

func (m *MockInternalClient) Batch(arg0 context.Context, arg1 *roachpb.BatchRequest, arg2 ...grpc.CallOption) (*roachpb.BatchResponse, error) {
	__antithesis_instrumentation__.Notify(176884)
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		__antithesis_instrumentation__.Notify(176886)
		varargs = append(varargs, a)
	}
	__antithesis_instrumentation__.Notify(176885)
	ret := m.ctrl.Call(m, "Batch", varargs...)
	ret0, _ := ret[0].(*roachpb.BatchResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockInternalClientMockRecorder) Batch(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	__antithesis_instrumentation__.Notify(176887)
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Batch", reflect.TypeOf((*MockInternalClient)(nil).Batch), varargs...)
}

func (m *MockInternalClient) GetAllSystemSpanConfigsThatApply(arg0 context.Context, arg1 *roachpb.GetAllSystemSpanConfigsThatApplyRequest, arg2 ...grpc.CallOption) (*roachpb.GetAllSystemSpanConfigsThatApplyResponse, error) {
	__antithesis_instrumentation__.Notify(176888)
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		__antithesis_instrumentation__.Notify(176890)
		varargs = append(varargs, a)
	}
	__antithesis_instrumentation__.Notify(176889)
	ret := m.ctrl.Call(m, "GetAllSystemSpanConfigsThatApply", varargs...)
	ret0, _ := ret[0].(*roachpb.GetAllSystemSpanConfigsThatApplyResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockInternalClientMockRecorder) GetAllSystemSpanConfigsThatApply(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	__antithesis_instrumentation__.Notify(176891)
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllSystemSpanConfigsThatApply", reflect.TypeOf((*MockInternalClient)(nil).GetAllSystemSpanConfigsThatApply), varargs...)
}

func (m *MockInternalClient) GetSpanConfigs(arg0 context.Context, arg1 *roachpb.GetSpanConfigsRequest, arg2 ...grpc.CallOption) (*roachpb.GetSpanConfigsResponse, error) {
	__antithesis_instrumentation__.Notify(176892)
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		__antithesis_instrumentation__.Notify(176894)
		varargs = append(varargs, a)
	}
	__antithesis_instrumentation__.Notify(176893)
	ret := m.ctrl.Call(m, "GetSpanConfigs", varargs...)
	ret0, _ := ret[0].(*roachpb.GetSpanConfigsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockInternalClientMockRecorder) GetSpanConfigs(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	__antithesis_instrumentation__.Notify(176895)
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSpanConfigs", reflect.TypeOf((*MockInternalClient)(nil).GetSpanConfigs), varargs...)
}

func (m *MockInternalClient) GossipSubscription(arg0 context.Context, arg1 *roachpb.GossipSubscriptionRequest, arg2 ...grpc.CallOption) (roachpb.Internal_GossipSubscriptionClient, error) {
	__antithesis_instrumentation__.Notify(176896)
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		__antithesis_instrumentation__.Notify(176898)
		varargs = append(varargs, a)
	}
	__antithesis_instrumentation__.Notify(176897)
	ret := m.ctrl.Call(m, "GossipSubscription", varargs...)
	ret0, _ := ret[0].(roachpb.Internal_GossipSubscriptionClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockInternalClientMockRecorder) GossipSubscription(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	__antithesis_instrumentation__.Notify(176899)
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GossipSubscription", reflect.TypeOf((*MockInternalClient)(nil).GossipSubscription), varargs...)
}

func (m *MockInternalClient) Join(arg0 context.Context, arg1 *roachpb.JoinNodeRequest, arg2 ...grpc.CallOption) (*roachpb.JoinNodeResponse, error) {
	__antithesis_instrumentation__.Notify(176900)
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		__antithesis_instrumentation__.Notify(176902)
		varargs = append(varargs, a)
	}
	__antithesis_instrumentation__.Notify(176901)
	ret := m.ctrl.Call(m, "Join", varargs...)
	ret0, _ := ret[0].(*roachpb.JoinNodeResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockInternalClientMockRecorder) Join(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	__antithesis_instrumentation__.Notify(176903)
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Join", reflect.TypeOf((*MockInternalClient)(nil).Join), varargs...)
}

func (m *MockInternalClient) RangeFeed(arg0 context.Context, arg1 *roachpb.RangeFeedRequest, arg2 ...grpc.CallOption) (roachpb.Internal_RangeFeedClient, error) {
	__antithesis_instrumentation__.Notify(176904)
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		__antithesis_instrumentation__.Notify(176906)
		varargs = append(varargs, a)
	}
	__antithesis_instrumentation__.Notify(176905)
	ret := m.ctrl.Call(m, "RangeFeed", varargs...)
	ret0, _ := ret[0].(roachpb.Internal_RangeFeedClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockInternalClientMockRecorder) RangeFeed(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	__antithesis_instrumentation__.Notify(176907)
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RangeFeed", reflect.TypeOf((*MockInternalClient)(nil).RangeFeed), varargs...)
}

func (m *MockInternalClient) RangeLookup(arg0 context.Context, arg1 *roachpb.RangeLookupRequest, arg2 ...grpc.CallOption) (*roachpb.RangeLookupResponse, error) {
	__antithesis_instrumentation__.Notify(176908)
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		__antithesis_instrumentation__.Notify(176910)
		varargs = append(varargs, a)
	}
	__antithesis_instrumentation__.Notify(176909)
	ret := m.ctrl.Call(m, "RangeLookup", varargs...)
	ret0, _ := ret[0].(*roachpb.RangeLookupResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockInternalClientMockRecorder) RangeLookup(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	__antithesis_instrumentation__.Notify(176911)
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RangeLookup", reflect.TypeOf((*MockInternalClient)(nil).RangeLookup), varargs...)
}

func (m *MockInternalClient) ResetQuorum(arg0 context.Context, arg1 *roachpb.ResetQuorumRequest, arg2 ...grpc.CallOption) (*roachpb.ResetQuorumResponse, error) {
	__antithesis_instrumentation__.Notify(176912)
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		__antithesis_instrumentation__.Notify(176914)
		varargs = append(varargs, a)
	}
	__antithesis_instrumentation__.Notify(176913)
	ret := m.ctrl.Call(m, "ResetQuorum", varargs...)
	ret0, _ := ret[0].(*roachpb.ResetQuorumResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockInternalClientMockRecorder) ResetQuorum(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	__antithesis_instrumentation__.Notify(176915)
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResetQuorum", reflect.TypeOf((*MockInternalClient)(nil).ResetQuorum), varargs...)
}

func (m *MockInternalClient) TenantSettings(arg0 context.Context, arg1 *roachpb.TenantSettingsRequest, arg2 ...grpc.CallOption) (roachpb.Internal_TenantSettingsClient, error) {
	__antithesis_instrumentation__.Notify(176916)
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		__antithesis_instrumentation__.Notify(176918)
		varargs = append(varargs, a)
	}
	__antithesis_instrumentation__.Notify(176917)
	ret := m.ctrl.Call(m, "TenantSettings", varargs...)
	ret0, _ := ret[0].(roachpb.Internal_TenantSettingsClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockInternalClientMockRecorder) TenantSettings(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	__antithesis_instrumentation__.Notify(176919)
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TenantSettings", reflect.TypeOf((*MockInternalClient)(nil).TenantSettings), varargs...)
}

func (m *MockInternalClient) TokenBucket(arg0 context.Context, arg1 *roachpb.TokenBucketRequest, arg2 ...grpc.CallOption) (*roachpb.TokenBucketResponse, error) {
	__antithesis_instrumentation__.Notify(176920)
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		__antithesis_instrumentation__.Notify(176922)
		varargs = append(varargs, a)
	}
	__antithesis_instrumentation__.Notify(176921)
	ret := m.ctrl.Call(m, "TokenBucket", varargs...)
	ret0, _ := ret[0].(*roachpb.TokenBucketResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockInternalClientMockRecorder) TokenBucket(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	__antithesis_instrumentation__.Notify(176923)
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TokenBucket", reflect.TypeOf((*MockInternalClient)(nil).TokenBucket), varargs...)
}

func (m *MockInternalClient) UpdateSpanConfigs(arg0 context.Context, arg1 *roachpb.UpdateSpanConfigsRequest, arg2 ...grpc.CallOption) (*roachpb.UpdateSpanConfigsResponse, error) {
	__antithesis_instrumentation__.Notify(176924)
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		__antithesis_instrumentation__.Notify(176926)
		varargs = append(varargs, a)
	}
	__antithesis_instrumentation__.Notify(176925)
	ret := m.ctrl.Call(m, "UpdateSpanConfigs", varargs...)
	ret0, _ := ret[0].(*roachpb.UpdateSpanConfigsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockInternalClientMockRecorder) UpdateSpanConfigs(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	__antithesis_instrumentation__.Notify(176927)
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateSpanConfigs", reflect.TypeOf((*MockInternalClient)(nil).UpdateSpanConfigs), varargs...)
}

type MockInternal_RangeFeedClient struct {
	ctrl     *gomock.Controller
	recorder *MockInternal_RangeFeedClientMockRecorder
}

type MockInternal_RangeFeedClientMockRecorder struct {
	mock *MockInternal_RangeFeedClient
}

func NewMockInternal_RangeFeedClient(ctrl *gomock.Controller) *MockInternal_RangeFeedClient {
	__antithesis_instrumentation__.Notify(176928)
	mock := &MockInternal_RangeFeedClient{ctrl: ctrl}
	mock.recorder = &MockInternal_RangeFeedClientMockRecorder{mock}
	return mock
}

func (m *MockInternal_RangeFeedClient) EXPECT() *MockInternal_RangeFeedClientMockRecorder {
	__antithesis_instrumentation__.Notify(176929)
	return m.recorder
}

func (m *MockInternal_RangeFeedClient) CloseSend() error {
	__antithesis_instrumentation__.Notify(176930)
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CloseSend")
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockInternal_RangeFeedClientMockRecorder) CloseSend() *gomock.Call {
	__antithesis_instrumentation__.Notify(176931)
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CloseSend", reflect.TypeOf((*MockInternal_RangeFeedClient)(nil).CloseSend))
}

func (m *MockInternal_RangeFeedClient) Context() context.Context {
	__antithesis_instrumentation__.Notify(176932)
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

func (mr *MockInternal_RangeFeedClientMockRecorder) Context() *gomock.Call {
	__antithesis_instrumentation__.Notify(176933)
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockInternal_RangeFeedClient)(nil).Context))
}

func (m *MockInternal_RangeFeedClient) Header() (metadata.MD, error) {
	__antithesis_instrumentation__.Notify(176934)
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Header")
	ret0, _ := ret[0].(metadata.MD)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockInternal_RangeFeedClientMockRecorder) Header() *gomock.Call {
	__antithesis_instrumentation__.Notify(176935)
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Header", reflect.TypeOf((*MockInternal_RangeFeedClient)(nil).Header))
}

func (m *MockInternal_RangeFeedClient) Recv() (*roachpb.RangeFeedEvent, error) {
	__antithesis_instrumentation__.Notify(176936)
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Recv")
	ret0, _ := ret[0].(*roachpb.RangeFeedEvent)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockInternal_RangeFeedClientMockRecorder) Recv() *gomock.Call {
	__antithesis_instrumentation__.Notify(176937)
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Recv", reflect.TypeOf((*MockInternal_RangeFeedClient)(nil).Recv))
}

func (m *MockInternal_RangeFeedClient) RecvMsg(arg0 interface{}) error {
	__antithesis_instrumentation__.Notify(176938)
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockInternal_RangeFeedClientMockRecorder) RecvMsg(arg0 interface{}) *gomock.Call {
	__antithesis_instrumentation__.Notify(176939)
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockInternal_RangeFeedClient)(nil).RecvMsg), arg0)
}

func (m *MockInternal_RangeFeedClient) SendMsg(arg0 interface{}) error {
	__antithesis_instrumentation__.Notify(176940)
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockInternal_RangeFeedClientMockRecorder) SendMsg(arg0 interface{}) *gomock.Call {
	__antithesis_instrumentation__.Notify(176941)
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockInternal_RangeFeedClient)(nil).SendMsg), arg0)
}

func (m *MockInternal_RangeFeedClient) Trailer() metadata.MD {
	__antithesis_instrumentation__.Notify(176942)
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Trailer")
	ret0, _ := ret[0].(metadata.MD)
	return ret0
}

func (mr *MockInternal_RangeFeedClientMockRecorder) Trailer() *gomock.Call {
	__antithesis_instrumentation__.Notify(176943)
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Trailer", reflect.TypeOf((*MockInternal_RangeFeedClient)(nil).Trailer))
}
