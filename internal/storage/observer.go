package storage

import (
	"context"
	"fmt"

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

func (base *Repo) UpdateStatus(ctx context.Context, order *model.Order, accrual *decimal.Decimal) (err error) {

	tx, err := base.master.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("can't update orders status, begin transaction error, %w", err)
	}

	_, err = tx.ExecContext(ctx, "UPDATE orders SET status = $1 WHERE onumber = $2", order.Status, order.Onumber)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("can't update orders status, %w", err)
	}

	if accrual != nil {
		_, err = tx.ExecContext(ctx, "UPDATE bonuses SET bonus_accrual = $1 WHERE onumber = $2", accrual, order.Onumber)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("can't update bonuses during update status %w", err)
		}
	}

	tx.Commit()

	return nil
}
