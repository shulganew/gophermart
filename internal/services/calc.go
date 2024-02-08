package services

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/shulganew/gophermart/internal/entities"
)

type CalculationService struct {
	stor CalcRepo
}

type CalcRepo interface {
	GetBonuses(ctx context.Context, userID uuid.UUID) (accrual decimal.Decimal, err error)
	GetWithdrawn(ctx context.Context, userID uuid.UUID) (accrual decimal.Decimal, err error)
	GetWithdrawals(ctx context.Context, userID uuid.UUID) (withdrawn decimal.Decimal, err error)
	Withdrawals(ctx context.Context, userID uuid.UUID) ([]entities.Withdrawals, error)
	IsPreOrder(ctx context.Context, userID uuid.UUID, order string) (isPreOrder bool, err error)
	MovePreOrder(ctx context.Context, order *entities.Order) (err error)
	SetAccrual(ctx context.Context, order string, accrual decimal.Decimal) (err error)
	MakeWithdrawn(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) error
}

func NewCalcService(stor CalcRepo) *CalculationService {
	return &CalculationService{stor: stor}
}

func (m *CalculationService) IsPreOrder(ctx context.Context, userID uuid.UUID, order string) (isPreOrder bool, err error) {
	return m.stor.IsPreOrder(ctx, userID, order)
}

// Make preorder (created with withdrawals) regular order.
func (m *CalculationService) MovePreOrder(ctx context.Context, order *entities.Order) (err error) {
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

func (m *CalculationService) GetBonuses(ctx context.Context, userID uuid.UUID) (bonuses decimal.Decimal, err error) {
	bonuses, err = m.stor.GetBonuses(ctx, userID)
	if err != nil {
		return decimal.Zero, fmt.Errorf("can't get user's bonuses: %w", err)
	}

	return
}

func (m *CalculationService) GetWithdrawn(ctx context.Context, userID uuid.UUID) (withdrawn decimal.Decimal, err error) {
	withdrawn, err = m.stor.GetWithdrawn(ctx, userID)
	if err != nil {
		return decimal.Zero, fmt.Errorf("can't get user's withdrawals: %w", err)
	}
	return
}

func (m *CalculationService) CheckBalance(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) (isEnough bool, err error) {
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

func (m *CalculationService) GetWithdrawals(ctx context.Context, userID uuid.UUID) (wds []entities.Withdrawals, err error) {
	wds, err = m.stor.Withdrawals(ctx, userID)
	return
}

// Move user's amount from bonuses to withdrawals.
func (m *CalculationService) MakeWithdrawn(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) (err error) {
	err = m.stor.MakeWithdrawn(ctx, userID, amount)
	return
}
