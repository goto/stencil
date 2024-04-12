// Code generated by mockery v2.42.1. DO NOT EDIT.

package mocks

import (
	context "context"

	changedetector "github.com/goto/stencil/core/changedetector"

	mock "github.com/stretchr/testify/mock"
)

// NotificationEventRepository is an autogenerated mock type for the NotificationEventRepository type
type NotificationEventRepository struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, event
func (_m *NotificationEventRepository) Create(ctx context.Context, event changedetector.NotificationEvent) (changedetector.NotificationEvent, error) {
	ret := _m.Called(ctx, event)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 changedetector.NotificationEvent
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, changedetector.NotificationEvent) (changedetector.NotificationEvent, error)); ok {
		return rf(ctx, event)
	}
	if rf, ok := ret.Get(0).(func(context.Context, changedetector.NotificationEvent) changedetector.NotificationEvent); ok {
		r0 = rf(ctx, event)
	} else {
		r0 = ret.Get(0).(changedetector.NotificationEvent)
	}

	if rf, ok := ret.Get(1).(func(context.Context, changedetector.NotificationEvent) error); ok {
		r1 = rf(ctx, event)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByNameSpaceSchemaAndVersionSuccess provides a mock function with given fields: ctx, namespace, schemaID, versionID, success
func (_m *NotificationEventRepository) GetByNameSpaceSchemaAndVersionSuccess(ctx context.Context, namespace string, schemaID int32, versionID string, success bool) (changedetector.NotificationEvent, error) {
	ret := _m.Called(ctx, namespace, schemaID, versionID, success)

	if len(ret) == 0 {
		panic("no return value specified for GetByNameSpaceSchemaAndVersionSuccess")
	}

	var r0 changedetector.NotificationEvent
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, int32, string, bool) (changedetector.NotificationEvent, error)); ok {
		return rf(ctx, namespace, schemaID, versionID, success)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, int32, string, bool) changedetector.NotificationEvent); ok {
		r0 = rf(ctx, namespace, schemaID, versionID, success)
	} else {
		r0 = ret.Get(0).(changedetector.NotificationEvent)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, int32, string, bool) error); ok {
		r1 = rf(ctx, namespace, schemaID, versionID, success)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, ID
func (_m *NotificationEventRepository) Update(ctx context.Context, ID string) (changedetector.NotificationEvent, error) {
	ret := _m.Called(ctx, ID)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 changedetector.NotificationEvent
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (changedetector.NotificationEvent, error)); ok {
		return rf(ctx, ID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) changedetector.NotificationEvent); ok {
		r0 = rf(ctx, ID)
	} else {
		r0 = ret.Get(0).(changedetector.NotificationEvent)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, ID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewNotificationEventRepository creates a new instance of NotificationEventRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewNotificationEventRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *NotificationEventRepository {
	mock := &NotificationEventRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
