// Code generated by mockery 2.9.0. DO NOT EDIT.

package mocks

import (
	context "context"

	"github.com/odpf/stencil/models"
	mock "github.com/stretchr/testify/mock"
)

// MetadataService is an autogenerated mock type for the MetadataService type
type MetadataService struct {
	mock.Mock
}

// Exists provides a mock function with given fields: _a0, _a1
func (_m *MetadataService) Exists(_a0 context.Context, _a1 *models.Snapshot) bool {
	ret := _m.Called(_a0, _a1)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, *models.Snapshot) bool); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// GetSnapshotByFields provides a mock function with given fields: _a0, _a1, _a2, _a3, _a4
func (_m *MetadataService) GetSnapshotByFields(_a0 context.Context, _a1 string, _a2 string, _a3 string, _a4 bool) (*models.Snapshot, error) {
	ret := _m.Called(_a0, _a1, _a2, _a3, _a4)

	var r0 *models.Snapshot
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, bool) *models.Snapshot); ok {
		r0 = rf(_a0, _a1, _a2, _a3, _a4)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Snapshot)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string, bool) error); ok {
		r1 = rf(_a0, _a1, _a2, _a3, _a4)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSnapshotByID provides a mock function with given fields: _a0, _a1
func (_m *MetadataService) GetSnapshotByID(_a0 context.Context, _a1 int64) (*models.Snapshot, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *models.Snapshot
	if rf, ok := ret.Get(0).(func(context.Context, int64) *models.Snapshot); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Snapshot)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: _a0, _a1
func (_m *MetadataService) List(_a0 context.Context, _a1 *models.Snapshot) ([]*models.Snapshot, error) {
	ret := _m.Called(_a0, _a1)

	var r0 []*models.Snapshot
	if rf, ok := ret.Get(0).(func(context.Context, *models.Snapshot) []*models.Snapshot); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.Snapshot)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *models.Snapshot) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateLatestVersion provides a mock function with given fields: _a0, _a1
func (_m *MetadataService) UpdateLatestVersion(_a0 context.Context, _a1 *models.Snapshot) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.Snapshot) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}