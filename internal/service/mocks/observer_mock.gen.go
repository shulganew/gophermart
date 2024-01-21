// Code generated by MockGen. DO NOT EDIT.
// Source: internal/services/observer.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	decimal "github.com/shopspring/decimal"
	model "github.com/shulganew/gophermart/internal/model"
)

// MockObserverUpdater is a mock of ObserverUpdater interface.
type MockObserverUpdater struct {
	ctrl     *gomock.Controller
	recorder *MockObserverUpdaterMockRecorder
}

// MockObserverUpdaterMockRecorder is the mock recorder for MockObserverUpdater.
type MockObserverUpdaterMockRecorder struct {
	mock *MockObserverUpdater
}

// NewMockObserverUpdater creates a new mock instance.
func NewMockObserverUpdater(ctrl *gomock.Controller) *MockObserverUpdater {
	mock := &MockObserverUpdater{ctrl: ctrl}
	mock.recorder = &MockObserverUpdaterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockObserverUpdater) EXPECT() *MockObserverUpdaterMockRecorder {
	return m.recorder
}

// LoadPocessing mocks base method.
func (m *MockObserverUpdater) LoadPocessing(ctx context.Context) ([]model.Order, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LoadPocessing", ctx)
	ret0, _ := ret[0].([]model.Order)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LoadPocessing indicates an expected call of LoadPocessing.
func (mr *MockObserverUpdaterMockRecorder) LoadPocessing(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoadPocessing", reflect.TypeOf((*MockObserverUpdater)(nil).LoadPocessing), ctx)
}

// UpdateStatus mocks base method.
func (m *MockObserverUpdater) UpdateStatus(ctx context.Context, order *model.Order, accrual *decimal.Decimal) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateStatus", ctx, order, accrual)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateStatus indicates an expected call of UpdateStatus.
func (mr *MockObserverUpdaterMockRecorder) UpdateStatus(ctx, order, accrual interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateStatus", reflect.TypeOf((*MockObserverUpdater)(nil).UpdateStatus), ctx, order, accrual)
}
