// Code generated by MockGen. DO NOT EDIT.
// Source: internal/services/market.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	uuid "github.com/gofrs/uuid"
	gomock "github.com/golang/mock/gomock"
	decimal "github.com/shopspring/decimal"
	model "github.com/shulganew/gophermart/internal/model"
)

// MockMarketPlaceholder is a mock of MarketPlaceholder interface.
type MockMarketPlaceholder struct {
	ctrl     *gomock.Controller
	recorder *MockMarketPlaceholderMockRecorder
}

// MockMarketPlaceholderMockRecorder is the mock recorder for MockMarketPlaceholder.
type MockMarketPlaceholderMockRecorder struct {
	mock *MockMarketPlaceholder
}

// NewMockMarketPlaceholder creates a new mock instance.
func NewMockMarketPlaceholder(ctrl *gomock.Controller) *MockMarketPlaceholder {
	mock := &MockMarketPlaceholder{ctrl: ctrl}
	mock.recorder = &MockMarketPlaceholderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMarketPlaceholder) EXPECT() *MockMarketPlaceholderMockRecorder {
	return m.recorder
}

// AddOrder mocks base method.
func (m *MockMarketPlaceholder) AddOrder(ctx context.Context, userID *uuid.UUID, order string, isPreorder bool, withdraw *decimal.Decimal) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddOrder", ctx, userID, order, isPreorder, withdraw)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddOrder indicates an expected call of AddOrder.
func (mr *MockMarketPlaceholderMockRecorder) AddOrder(ctx, userID, order, isPreorder, withdraw interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddOrder", reflect.TypeOf((*MockMarketPlaceholder)(nil).AddOrder), ctx, userID, order, isPreorder, withdraw)
}

// GetAccruals mocks base method.
func (m *MockMarketPlaceholder) GetAccruals(ctx context.Context, userID *uuid.UUID) (*decimal.Decimal, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccruals", ctx, userID)
	ret0, _ := ret[0].(*decimal.Decimal)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAccruals indicates an expected call of GetAccruals.
func (mr *MockMarketPlaceholderMockRecorder) GetAccruals(ctx, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccruals", reflect.TypeOf((*MockMarketPlaceholder)(nil).GetAccruals), ctx, userID)
}

// GetOrders mocks base method.
func (m *MockMarketPlaceholder) GetOrders(ctx context.Context, userID *uuid.UUID) ([]model.Order, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrders", ctx, userID)
	ret0, _ := ret[0].([]model.Order)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrders indicates an expected call of GetOrders.
func (mr *MockMarketPlaceholderMockRecorder) GetOrders(ctx, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrders", reflect.TypeOf((*MockMarketPlaceholder)(nil).GetOrders), ctx, userID)
}

// GetWithdrawns mocks base method.
func (m *MockMarketPlaceholder) GetWithdrawns(ctx context.Context, userID *uuid.UUID) (*decimal.Decimal, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWithdrawns", ctx, userID)
	ret0, _ := ret[0].(*decimal.Decimal)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetWithdrawns indicates an expected call of GetWithdrawns.
func (mr *MockMarketPlaceholderMockRecorder) GetWithdrawns(ctx, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWithdrawns", reflect.TypeOf((*MockMarketPlaceholder)(nil).GetWithdrawns), ctx, userID)
}

// IsExistForOtherUsers mocks base method.
func (m *MockMarketPlaceholder) IsExistForOtherUsers(ctx context.Context, userID *uuid.UUID, order string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsExistForOtherUsers", ctx, userID, order)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsExistForOtherUsers indicates an expected call of IsExistForOtherUsers.
func (mr *MockMarketPlaceholderMockRecorder) IsExistForOtherUsers(ctx, userID, order interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsExistForOtherUsers", reflect.TypeOf((*MockMarketPlaceholder)(nil).IsExistForOtherUsers), ctx, userID, order)
}

// IsExistForUser mocks base method.
func (m *MockMarketPlaceholder) IsExistForUser(ctx context.Context, userID *uuid.UUID, order string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsExistForUser", ctx, userID, order)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsExistForUser indicates an expected call of IsExistForUser.
func (mr *MockMarketPlaceholderMockRecorder) IsExistForUser(ctx, userID, order interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsExistForUser", reflect.TypeOf((*MockMarketPlaceholder)(nil).IsExistForUser), ctx, userID, order)
}

// IsPreOrder mocks base method.
func (m *MockMarketPlaceholder) IsPreOrder(ctx context.Context, userID *uuid.UUID, order string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsPreOrder", ctx, userID, order)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsPreOrder indicates an expected call of IsPreOrder.
func (mr *MockMarketPlaceholderMockRecorder) IsPreOrder(ctx, userID, order interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsPreOrder", reflect.TypeOf((*MockMarketPlaceholder)(nil).IsPreOrder), ctx, userID, order)
}

// MovePreOrder mocks base method.
func (m *MockMarketPlaceholder) MovePreOrder(ctx context.Context, order *model.Order) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MovePreOrder", ctx, order)
	ret0, _ := ret[0].(error)
	return ret0
}

// MovePreOrder indicates an expected call of MovePreOrder.
func (mr *MockMarketPlaceholderMockRecorder) MovePreOrder(ctx, order interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MovePreOrder", reflect.TypeOf((*MockMarketPlaceholder)(nil).MovePreOrder), ctx, order)
}

// Withdrawals mocks base method.
func (m *MockMarketPlaceholder) Withdrawals(ctx context.Context, userID *uuid.UUID) ([]model.Withdrawals, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Withdrawals", ctx, userID)
	ret0, _ := ret[0].([]model.Withdrawals)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Withdrawals indicates an expected call of Withdrawals.
func (mr *MockMarketPlaceholderMockRecorder) Withdrawals(ctx, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Withdrawals", reflect.TypeOf((*MockMarketPlaceholder)(nil).Withdrawals), ctx, userID)
}
