package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/shulganew/gophermart/internal/model"
)

// User registration service
type Market struct {
	stor MarketPlaceholder
}

type MarketPlaceholder interface {
	GetOrders(ctx context.Context, userID *uuid.UUID) ([]model.Order, error)
	AddOrder(ctx context.Context, userID *uuid.UUID, order string, isPreorder bool, withdraw *decimal.Decimal) error
	IsExistForUser(ctx context.Context, userID *uuid.UUID, order string) (isExist bool, err error)
	IsExistForOtherUsers(ctx context.Context, userID *uuid.UUID, order string) (isExist bool, err error)
	GetAccruals(ctx context.Context, userID *uuid.UUID) (accrual *decimal.Decimal, err error)
	GetWithdrawns(ctx context.Context, userID *uuid.UUID) (withdrawn *decimal.Decimal, err error)
	Withdrawals(ctx context.Context, userID *uuid.UUID) ([]model.Withdrawals, error)
	IsPreOrder(ctx context.Context, userID *uuid.UUID, order string) (isPreOrder bool, err error)
	MovePreOrder(ctx context.Context, order *model.Order) (err error)
}

func NewMarket(stor MarketPlaceholder) *Market {
	return &Market{stor: stor}
}

func (m *Market) AddOrder(ctx context.Context, isPreOrder bool, order *model.Order) (existed bool, err error) {
	// Add order to the database.
	err = m.stor.AddOrder(ctx, order.UserID, order.Onumber, isPreOrder, order.Bonus.Used)
	if err != nil {
		var pgErr *pq.Error
		// If Order exist in the DataBase
		if errors.As(err, &pgErr) && pgerrcode.UniqueViolation == pgErr.Code {
			return true, err
		}
		return false, fmt.Errorf("error during add order: %w", err)
	}

	return false, nil
}

func (m *Market) GetOrders(ctx context.Context, userID *uuid.UUID) (orders []model.Order, err error) {

	orders, err = m.stor.GetOrders(ctx, userID)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (m *Market) IsPreOrder(ctx context.Context, userID *uuid.UUID, order string) (isPreOrder bool, err error) {
	return m.stor.IsPreOrder(ctx, userID, order)
}

func (m *Market) MovePreOrder(ctx context.Context, order *model.Order) (err error) {
	return m.stor.MovePreOrder(ctx, order)
}

func (m *Market) IsExistForUser(ctx context.Context, userID *uuid.UUID, order string) (isExist bool, err error) {
	return m.stor.IsExistForUser(ctx, userID, order)
}

func (m *Market) IsExistForOtherUsers(ctx context.Context, userID *uuid.UUID, order string) (isExist bool, err error) {
	return m.stor.IsExistForOtherUsers(ctx, userID, order)
}

func (m *Market) GetBalance(ctx context.Context, userID *uuid.UUID) (acc *decimal.Decimal, withdrawn *decimal.Decimal, err error) {
	acc, err = m.stor.GetAccruals(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	withdrawn, err = m.stor.GetWithdrawns(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	return
}

func (m *Market) CheckBalance(ctx context.Context, userID *uuid.UUID, amount *decimal.Decimal) (isEnough bool, err error) {
	acc, err := m.stor.GetAccruals(ctx, userID)
	if err != nil {
		return false, err
	}

	wd, err := m.stor.GetWithdrawns(ctx, userID)
	if err != nil {
		return false, err
	}

	bonuses := acc.Sub(*wd)
	rest := bonuses.Sub(*amount)

	if rest.IsNegative() {
		return false, nil
	}
	return true, nil
}

func (m *Market) GetWithdrawals(ctx context.Context, userID *uuid.UUID) (wds []model.Withdrawals, err error) {
	wds, err = m.stor.Withdrawals(ctx, userID)
	return
}
