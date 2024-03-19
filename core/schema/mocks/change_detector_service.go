// Code generated by mockery v2.42.1. DO NOT EDIT.

package mocks

import (
	changedetector "github.com/goto/stencil/core/changedetector"
	mock "github.com/stretchr/testify/mock"

	stencilv1beta1 "github.com/goto/stencil/proto/gotocompany/stencil/v1beta1"
)

// ChangeDetectorService is an autogenerated mock type for the ChangeDetectorService type
type ChangeDetectorService struct {
	mock.Mock
}

// IdentifySchemaChange provides a mock function with given fields: request
func (_m *ChangeDetectorService) IdentifySchemaChange(request *changedetector.ChangeRequest) (*stencilv1beta1.SchemaChangedEvent, error) {
	ret := _m.Called(request)

	if len(ret) == 0 {
		panic("no return value specified for IdentifySchemaChange")
	}

	var r0 *stencilv1beta1.SchemaChangedEvent
	var r1 error
	if rf, ok := ret.Get(0).(func(*changedetector.ChangeRequest) (*stencilv1beta1.SchemaChangedEvent, error)); ok {
		return rf(request)
	}
	if rf, ok := ret.Get(0).(func(*changedetector.ChangeRequest) *stencilv1beta1.SchemaChangedEvent); ok {
		r0 = rf(request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*stencilv1beta1.SchemaChangedEvent)
		}
	}

	if rf, ok := ret.Get(1).(func(*changedetector.ChangeRequest) error); ok {
		r1 = rf(request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewChangeDetectorService creates a new instance of ChangeDetectorService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewChangeDetectorService(t interface {
	mock.TestingT
	Cleanup(func())
}) *ChangeDetectorService {
	mock := &ChangeDetectorService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}