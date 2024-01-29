package storage

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/shulganew/gophermart/internal/model"
)

// Load all orders with not finished preparation status.
func (base *Repo) LoadPocessing(ctx context.Context) ([]model.Order, error) {
	query := `
	SELECT users.user_id, orders.onumber, orders.uploaded, orders.status
	FROM orders INNER JOIN users ON orders.user_id = users.user_id 
	WHERE (status = 'NEW' OR status = 'REGISTERED' OR status = 'PROCESSING') AND is_preorder = FALSE
	`

	rows, err := base.master.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	orders := []model.Order{}
	for rows.Next() {
		var order model.Order
		var status string
		err = rows.Scan(&order.UserID, &order.Onumber, &order.Uploaded, &status)
		if err != nil {
			return nil, err
		}
		order.Status = model.Status(status)
		orders = append(orders, order)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (base *Repo) UpdateStatus(ctx context.Context, order string, status model.Status) (err error) {

	if err != nil {
		return fmt.Errorf("can't update orders status, begin transaction error, %w", err)
	}

	_, err = base.master.ExecContext(ctx, "UPDATE orders SET status = $1 WHERE onumber = $2", status, order)
	if err != nil {
		return fmt.Errorf("can't update orders status, %w", err)
	}

	return nil
}

func (base *Repo) SetAccrual(ctx context.Context, order string, accrual decimal.Decimal) (err error) {

	if accrual != decimal.Zero {
		_, err = base.master.ExecContext(ctx, "UPDATE orders SET accrual = $1 WHERE onumber = $2", order, accrual)
		if err != nil {

			return fmt.Errorf("can't update order's accrual during update status %w", err)
		}
	}

	return nil
}

// Calculate total user's bonuses to attribute users.bonuses
func (base *Repo) CalculateBonuses(ctx context.Context, userID *uuid.UUID) (accrual *decimal.Decimal, err error) {
	query := `
	SELECT SUM(bonuses.bonus_accrual)
		FROM orders 
		INNER JOIN users ON orders.user_id = users.user_id
		INNER JOIN bonuses ON orders.onumber = bonuses.onumber;
	`

	row := base.master.QueryRowContext(ctx, query)

	err = row.Scan(&accrual)
	if err != nil {
		return nil, err
	}

	// If Postgres SUM is 0, it return nil
	if accrual == nil {
		return &decimal.Zero, nil
	}

	return accrual, nil
}

func (base *Repo) CalculateWithdrawns(ctx context.Context, userID *uuid.UUID) (withdrawn *decimal.Decimal, err error) {
	query := `
	SELECT SUM(bonuses.bonus_withdrawn)
		FROM orders 
		INNER JOIN users ON orders.user_id = users.user_id
		INNER JOIN bonuses ON orders.onumber = bonuses.onumber;
	`

	row := base.master.QueryRowContext(ctx, query)

	err = row.Scan(&withdrawn)
	if err != nil {
		return nil, err
	}

	return withdrawn, nil
}
