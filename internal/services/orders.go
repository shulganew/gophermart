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

type OrderService struct {
	stor OrderRepo
}

type OrderRepo interface {
	AddOrder(ctx context.Context, userID uuid.UUID, order string, isPreorder bool, withdraw decimal.Decimal) error
	GetOrders(ctx context.Context, userID uuid.UUID) ([]model.Order, error)
	IsExistForUser(ctx context.Context, userID uuid.UUID, order string) (isExist bool, err error)
}

func NewOrderService(stor OrderRepo) *OrderService {
	return &OrderService{stor: stor}
}

func (m *OrderService) AddOrder(ctx context.Context, isPreOrder bool, order *model.Order) (existed bool, err error) {
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

func (m *OrderService) GetOrders(ctx context.Context, userID uuid.UUID) (orders []model.Order, err error) {

	orders, err = m.stor.GetOrders(ctx, userID)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (m *OrderService) IsExistForUser(ctx context.Context, userID uuid.UUID, order string) (isExist bool, err error) {
	return m.stor.IsExistForUser(ctx, userID, order)
}