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
	AddOrder(ctx context.Context, userID uuid.UUID, order string, isPreorder bool, withdraw decimal.Decimal) error
	GetOrders(ctx context.Context, userID uuid.UUID) ([]model.Order, error)
	IsExistForUser(ctx context.Context, userID uuid.UUID, order string) (isExist bool, err error)
	GetBonuses(ctx context.Context, userID uuid.UUID) (accrual decimal.Decimal, err error)
	GetWithdrawn(ctx context.Context, userID uuid.UUID) (accrual decimal.Decimal, err error)
	GetWithdrawals(ctx context.Context, userID uuid.UUID) (withdrawn decimal.Decimal, err error)
	Withdrawals(ctx context.Context, userID uuid.UUID) ([]model.Withdrawals, error)
	IsPreOrder(ctx context.Context, userID uuid.UUID, order string) (isPreOrder bool, err error)
	MovePreOrder(ctx context.Context, order *model.Order) (err error)
	SetAccrual(ctx context.Context, order string, accrual decimal.Decimal) (err error)
	MakeWithdrawn(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) error
}

func NewMarket(stor MarketPlaceholder) *Market {
	return &Market{stor: stor}
}

func (m *Market) AddOrder(ctx context.Context, isPreOrder bool, order *model.Order) (existed bool, err error) {
	// Add order to the database.
	err = m.stor.AddOrder(ctx, order.UserID, order.OrderNr, isPreOrder, order.Withdrawn)
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

func (m *Market) GetOrders(ctx context.Context, userID uuid.UUID) (orders []model.Order, err error) {

	orders, err = m.stor.GetOrders(ctx, userID)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (m *Market) IsPreOrder(ctx context.Context, userID uuid.UUID, order string) (isPreOrder bool, err error) {
	return m.stor.IsPreOrder(ctx, userID, order)
}

// Make preorder (created with withdrawals) regular order
func (m *Market) MovePreOrder(ctx context.Context, order *model.Order) (err error) {
	err = m.stor.MovePreOrder(ctx, order)
	if err != nil {
		return fmt.Errorf("can't move preOrder to Oreder: %w", err)
	}
	// if order has accruals
	if order.Accrual != decimal.Zero {
		err = m.stor.SetAccrual(ctx, order.OrderNr, order.Accrual)
		if err != nil {
			return fmt.Errorf("can't set accruals to preorder: %w", err)
		}
	}
	return
}

func (m *Market) IsExistForUser(ctx context.Context, userID uuid.UUID, order string) (isExist bool, err error) {
	return m.stor.IsExistForUser(ctx, userID, order)
}

func (m *Market) GetBonuses(ctx context.Context, userID uuid.UUID) (bonuses decimal.Decimal, err error) {
	bonuses, err = m.stor.GetBonuses(ctx, userID)
	if err != nil {
		return decimal.Zero, fmt.Errorf("can't get user's bonuses: %w", err)
	}

	return
}

func (m *Market) GetWithdrawn(ctx context.Context, userID uuid.UUID) (withdrawn decimal.Decimal, err error) {

	withdrawn, err = m.stor.GetWithdrawn(ctx, userID)
	if err != nil {
		return decimal.Zero, fmt.Errorf("can't get user's withdrawals: %w", err)
	}
	return
}

func (m *Market) CheckBalance(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) (isEnough bool, err error) {
	bonuses, err := m.stor.GetBonuses(ctx, userID)
	if err != nil {
		return false, err
	}

	saldo := bonuses.Sub(amount)

	if saldo.IsNegative() {
		return false, nil
	}
	return true, nil
}

func (m *Market) GetWithdrawals(ctx context.Context, userID uuid.UUID) (wds []model.Withdrawals, err error) {
	wds, err = m.stor.Withdrawals(ctx, userID)
	return
}

// Move user's amount from bonuses to withdrawals
func (m *Market) MakeWithdrawn(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) (err error) {
	err = m.stor.MakeWithdrawn(ctx, userID, amount)
	return
}
