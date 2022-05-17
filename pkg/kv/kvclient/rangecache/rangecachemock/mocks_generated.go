// Package rangecachemock is a generated GoMock package.
package rangecachemock

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	context "context"
	reflect "reflect"

	roachpb "github.com/cockroachdb/cockroach/pkg/roachpb"
	gomock "github.com/golang/mock/gomock"
)

type MockRangeDescriptorDB struct {
	ctrl     *gomock.Controller
	recorder *MockRangeDescriptorDBMockRecorder
}

type MockRangeDescriptorDBMockRecorder struct {
	mock *MockRangeDescriptorDB
}

func NewMockRangeDescriptorDB(ctrl *gomock.Controller) *MockRangeDescriptorDB {
	__antithesis_instrumentation__.Notify(89629)
	mock := &MockRangeDescriptorDB{ctrl: ctrl}
	mock.recorder = &MockRangeDescriptorDBMockRecorder{mock}
	return mock
}

func (m *MockRangeDescriptorDB) EXPECT() *MockRangeDescriptorDBMockRecorder {
	__antithesis_instrumentation__.Notify(89630)
	return m.recorder
}

func (m *MockRangeDescriptorDB) FirstRange() (*roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(89631)
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FirstRange")
	ret0, _ := ret[0].(*roachpb.RangeDescriptor)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockRangeDescriptorDBMockRecorder) FirstRange() *gomock.Call {
	__antithesis_instrumentation__.Notify(89632)
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FirstRange", reflect.TypeOf((*MockRangeDescriptorDB)(nil).FirstRange))
}

func (m *MockRangeDescriptorDB) RangeLookup(arg0 context.Context, arg1 roachpb.RKey, arg2 bool) ([]roachpb.RangeDescriptor, []roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(89633)
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RangeLookup", arg0, arg1, arg2)
	ret0, _ := ret[0].([]roachpb.RangeDescriptor)
	ret1, _ := ret[1].([]roachpb.RangeDescriptor)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (mr *MockRangeDescriptorDBMockRecorder) RangeLookup(arg0, arg1, arg2 interface{}) *gomock.Call {
	__antithesis_instrumentation__.Notify(89634)
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RangeLookup", reflect.TypeOf((*MockRangeDescriptorDB)(nil).RangeLookup), arg0, arg1, arg2)
}
