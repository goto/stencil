// Code generated by mockery v2.42.1. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"

	time "time"
)

// Producer is an autogenerated mock type for the Producer type
type Producer struct {
	mock.Mock
}

// PushMessagesWithRetries provides a mock function with given fields: topic, protoMessage, retries, retryInterval
func (_m *Producer) PushMessagesWithRetries(topic string, protoMessage protoreflect.ProtoMessage, retries int, retryInterval time.Duration) error {
	ret := _m.Called(topic, protoMessage, retries, retryInterval)

	if len(ret) == 0 {
		panic("no return value specified for PushMessagesWithRetries")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, protoreflect.ProtoMessage, int, time.Duration) error); ok {
		r0 = rf(topic, protoMessage, retries, retryInterval)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewProducer creates a new instance of Producer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewProducer(t interface {
	mock.TestingT
	Cleanup(func())
}) *Producer {
	mock := &Producer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
