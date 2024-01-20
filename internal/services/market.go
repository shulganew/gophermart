package services

import (
	"context"
	"errors"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
	"github.com/shulganew/gophermart/internal/model"
	"go.uber.org/zap"
)

// User registration service
type Market struct {
	stor MarketPlaceholder
}

type MarketPlaceholder interface {
	GetOrders(ctx context.Context, userID *uuid.UUID) ([]model.Order, error)
	SetOrder(ctx context.Context, userID *uuid.UUID, order string) error
	IsExistForUser(ctx context.Context, userID *uuid.UUID, order string) (isExist bool, err error)
	IsExistForOtherUsers(ctx context.Context, userID *uuid.UUID, order string) (isExist bool, err error)
	GetAccruals(ctx context.Context, userID *uuid.UUID) (accrual *decimal.Decimal, err error)
	GetWithdrawns(ctx context.Context, userID *uuid.UUID) (withdrawn *decimal.Decimal, err error)
	Withdrow(ctx context.Context, userID *uuid.UUID, order string, amount *decimal.Decimal) error
	Withdrawals(ctx context.Context, userID *uuid.UUID) ([]model.Withdrawals, error)
}

func NewMarket(stor MarketPlaceholder) *Market {
	return &Market{stor: stor}
}

func (m *Market) SetOrder(ctx context.Context, order *model.Order) (existed bool, err error) {

	// Add order to the database.
	err = m.stor.SetOrder(ctx, order.UserID, order.Onumber)
	if err != nil {
		var pgErr *pgconn.PgError
		// If Order exist in the DataBase
		if errors.As(err, &pgErr) && pgerrcode.UniqueViolation == pgErr.Code {
			zap.S().Infoln("Order exist: ", order)
			return true, err
		}
		zap.S().Errorln("Set order error: ", order)
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

	bonuses := acc.Sub(*amount)
	if bonuses.IsNegative() {
		return false, nil
	}
	return true, nil
}

func (m *Market) Withdrow(ctx context.Context, userID *uuid.UUID, order string, amount *decimal.Decimal) error {
	err := m.stor.Withdrow(ctx, userID, order, amount)
	return err
}

func (m *Market) GetWithdrawals(ctx context.Context, userID *uuid.UUID) (wds []model.Withdrawals, err error) {
	wds, err = m.stor.Withdrawals(ctx, userID)
	return
}
