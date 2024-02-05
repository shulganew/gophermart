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
	IsExistForOtherUser(ctx context.Context, userID uuid.UUID, order string) (isExist bool, err error)
	IsExist(ctx context.Context, order string) (isExist bool, err error)
}

func NewOrderService(stor OrderRepo) *OrderService {
	return &OrderService{stor: stor}
}

func (m *OrderService) AddOrder(ctx context.Context, isPreOrder bool, order *model.Order) (err error) {
	// Add order to the database.
	err = m.stor.AddOrder(ctx, order.UserID, order.OrderNr, isPreOrder, order.Withdrawn)
	if err != nil {
		var pgErr *pq.Error
		// If Order exist in the DataBase
		if errors.As(err, &pgErr) && pgerrcode.UniqueViolation == pgErr.Code {
			return err
		}
		return fmt.Errorf("error during add order: %w", err)
	}

	return nil
}

func (m *OrderService) GetOrders(ctx context.Context, userID uuid.UUID) (orders []model.Order, err error) {

	orders, err = m.stor.GetOrders(ctx, userID)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (m *OrderService) IsExistForOtherUser(ctx context.Context, userID uuid.UUID, order string) (isExist bool, err error) {
	return m.stor.IsExistForOtherUser(ctx, userID, order)
}

func (m *OrderService) IsExist(ctx context.Context, order string) (isExist bool, err error) {
	return m.stor.IsExist(ctx, order)
}
