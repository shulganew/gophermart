package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"github.com/shulganew/gophermart/internal/entities"
)

type OrderService struct {
	stor orderRepo
}

type orderRepo interface {
	AddOrder(ctx context.Context, data *entities.AddOrder) error
	GetOrders(ctx context.Context, userID uuid.UUID) ([]entities.Order, error)
	IsExistForOtherUser(ctx context.Context, userID uuid.UUID, order string) (isExist bool, err error)
	IsExist(ctx context.Context, order string) (isExist bool, err error)
}

func NewOrderService(stor orderRepo) *OrderService {
	return &OrderService{stor: stor}
}

func (m *OrderService) AddOrder(ctx context.Context, isPreOrder bool, order *entities.Order) (err error) {
	// Add order to the database.
	addOrder := entities.NewAddOrder(order.UserID.String(), order.OrderNr, isPreOrder, order.Withdrawn)
	err = m.stor.AddOrder(ctx, addOrder)
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

func (m *OrderService) GetOrders(ctx context.Context, userID uuid.UUID) (orders []entities.Order, err error) {
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
