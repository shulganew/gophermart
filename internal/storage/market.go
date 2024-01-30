package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/shulganew/gophermart/internal/model"
)

func (base *Repo) AddOrder(ctx context.Context, userID *uuid.UUID, order string, isPreOrder bool, withdrawn decimal.Decimal) error {

	_, err := base.master.ExecContext(ctx, "INSERT INTO orders (user_id, onumber, is_preorder, uploaded, withdrawn) VALUES ($1, $2, $3, $4, $5)", userID, order, isPreOrder, time.Now(), withdrawn)
	if err != nil {
		var pgErr *pq.Error
		// if order exist in DataBase
		if errors.As(err, &pgErr) && pgerrcode.UniqueViolation == pgErr.Code {
			return pgErr
		}
		return fmt.Errorf("error during set order to Storage, error: %w", err)
	}

	return nil
}

func (base *Repo) GetOrders(ctx context.Context, userID *uuid.UUID) ([]model.Order, error) {
	query := `
	SELECT user_id, onumber, uploaded, status, withdrawn, accrual
		FROM orders 
		WHERE is_preorder = FALSE AND user_id = $1
		ORDER BY uploaded DESC
		`
	orders := []model.Order{}
	err := base.master.MustBegin().SelectContext(ctx, &orders, query, userID)
	if err != nil {
		return nil, err
	}
	return orders, nil

}

func (base *Repo) IsExistForUser(ctx context.Context, userID *uuid.UUID, order string) (isExist bool, err error) {

	var ordersn int
	err = base.master.GetContext(ctx, &ordersn, "SELECT count(*) FROM orders WHERE user_id = $1 AND onumber = $2", userID, order)
	if err != nil {
		return true, fmt.Errorf("error during order search for user: %w", err)
	}
	return ordersn != 0, nil
}

func (base *Repo) IsExistForOtherUsers(ctx context.Context, userID *uuid.UUID, order string) (isExist bool, err error) {

	var ordersn int
	err = base.master.GetContext(ctx, &ordersn, "SELECT count(*) FROM orders WHERE user_id != $1 AND onumber = $2", userID, order)
	if err != nil {
		return true, fmt.Errorf("error during order search for user: %w", err)
	}

	return ordersn != 0, nil
}

func (base *Repo) GetWithdrawals(ctx context.Context, userID *uuid.UUID) (withdrawn decimal.Decimal, err error) {

	err = base.master.GetContext(ctx, &withdrawn, "SELECT withdrawals FROM users where user_id = $1", userID)
	if err != nil {
		return decimal.Zero, err
	}
	return withdrawn, nil
}

func (base *Repo) Withdrawals(ctx context.Context, userID *uuid.UUID) (wds []model.Withdrawals, err error) {

	query := `
	SELECT  onumber, withdrawn, uploaded
		FROM orders 
		WHERE user_id = $1
		ORDER BY uploaded DESC
		`

	err = base.master.MustBegin().SelectContext(ctx, &wds, query, userID)
	if err != nil {
		return nil, err
	}

	return
}

func (base *Repo) IsPreOrder(ctx context.Context, userID *uuid.UUID, order string) (bool, error) {
	query := `
	SELECT count(onumber)
	FROM orders
	WHERE user_id = $1 AND onumber = $2 AND is_preorder = TRUE
	`

	var is int
	err := base.master.GetContext(ctx, &is, query, userID, order)
	if err != nil {
		return true, fmt.Errorf("error during is preorder checking: %w", err)
	}

	return is != 0, err
}

// Move preorder to regular order. Add accruals for this order.
func (base *Repo) MovePreOrder(ctx context.Context, order *model.Order) (err error) {

	_, err = base.master.ExecContext(ctx, "UPDATE orders SET status = $1, is_preorder = $2 WHERE onumber = $3", order.Status, order.IsPreOrder, order.Onumber)
	if err != nil {
		return fmt.Errorf("move order error, can't move preoreder to order, %w", err)
	}

	return

}
