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

func (base *Repo) GetOrders(ctx context.Context, userID *uuid.UUID) ([]model.Order, error) {
	query := `
	SELECT users.user_id, orders.onumber, orders.uploaded, orders.status, bonuses.bonus_used, bonuses.bonus_accrual
		FROM orders 
		INNER JOIN users ON orders.user_id = users.user_id
		INNER JOIN bonuses ON orders.onumber = bonuses.onumber
		WHERE is_preorder = FALSE AND orders.user_id = $1
		ORDER BY orders.uploaded DESC
		`

	rows, err := base.master.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	orders := []model.Order{}
	for rows.Next() {
		var order model.Order
		var status string
		var used decimal.Decimal
		var accrual decimal.Decimal

		err = rows.Scan(&order.UserID, &order.Onumber, &order.Uploaded, &status, &used, &accrual)
		if err != nil {
			return nil, err
		}
		order.Status = model.Status(status)
		order.Bonus = model.NewBonus(&used, &accrual)
		orders = append(orders, order)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return orders, nil

}

func (base *Repo) AddOrder(ctx context.Context, userID *uuid.UUID, order string, isPreOrder bool, withdrawn *decimal.Decimal) error {

	_, err := base.master.ExecContext(ctx, "INSERT INTO orders (user_id, onumber, is_preorder, uploaded) VALUES ($1, $2, $3, $4)", userID, order, isPreOrder, time.Now())
	if err != nil {
		var pgErr *pq.Error
		// if order exist in DataBase
		if errors.As(err, &pgErr) && pgerrcode.UniqueViolation == pgErr.Code {
			return pgErr
		}
		return fmt.Errorf("error during set order to Storage, error: %w", err)
	}

	_, err = base.master.ExecContext(ctx, "INSERT INTO bonuses (onumber, bonus_used) VALUES ($1, $2)", order, *withdrawn)
	if err != nil {
		return fmt.Errorf("error during add bonuses order to Storage, error: %w", err)
	}

	return nil
}

func (base *Repo) IsExistForUser(ctx context.Context, userID *uuid.UUID, order string) (isExist bool, err error) {

	var ordersn int
	row := base.master.QueryRowContext(ctx, "SELECT count(*) FROM orders WHERE user_id = $1 AND onumber = $2", userID, order)
	err = row.Scan(&ordersn)
	if err != nil {
		return true, fmt.Errorf("error during order search for user: %w", err)
	}
	if ordersn == 0 {
		return false, nil
	}

	return true, nil
}

func (base *Repo) IsExistForOtherUsers(ctx context.Context, userID *uuid.UUID, order string) (isExist bool, err error) {
	var ordersn int
	row := base.master.QueryRowContext(ctx, "SELECT count(*) FROM orders WHERE user_id != $1 AND onumber = $2", userID, order)
	err = row.Scan(&ordersn)
	if err != nil {
		return true, fmt.Errorf("error during order search for user: %w", err)
	}
	if ordersn == 0 {
		return false, nil
	}

	return true, nil
}

func (base *Repo) GetAccruals(ctx context.Context, userID *uuid.UUID) (accrual *decimal.Decimal, err error) {
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

func (base *Repo) GetWithdrawns(ctx context.Context, userID *uuid.UUID) (withdrawn *decimal.Decimal, err error) {
	query := `
	SELECT SUM(bonuses.bonus_used)
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

func (base *Repo) Withdrawals(ctx context.Context, userID *uuid.UUID) (wds []model.Withdrawals, err error) {
	query := `
	SELECT  orders.onumber, bonuses.bonus_used, orders.uploaded
		FROM orders 
		INNER JOIN users ON orders.user_id = users.user_id
		INNER JOIN bonuses ON orders.onumber = bonuses.onumber
		ORDER BY orders.uploaded DESC
		`

	rows, err := base.master.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var wd model.Withdrawals
		var updloded time.Time
		err = rows.Scan(&wd.Onumber, &wd.Withdrawn, &updloded)
		if err != nil {
			return nil, err
		}
		wd.Uploaded = updloded.Format(time.RFC3339)
		wds = append(wds, wd)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return
}

func (base *Repo) IsPreOrder(ctx context.Context, userID *uuid.UUID, order string) (bool, error) {
	query := `
	SELECT count(orders.onumber)
	FROM orders INNER JOIN users ON orders.user_id = users.user_id 
	WHERE users.user_id = $1 AND orders.onumber = $2 AND is_preorder = TRUE
	`
	row := base.master.QueryRowContext(ctx, query, userID, order)
	var n int
	err := row.Scan(&n)

	return n == 1, err
}

// Move preorder to regular order. Add accruals for this order.
func (base *Repo) MovePreOrder(ctx context.Context, order *model.Order) (err error) {

	tx, err := base.master.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("move order error, begin transaction error, %w", err)
	}

	_, err = tx.ExecContext(ctx, "UPDATE orders SET status = $1, is_preorder = $2 WHERE onumber = $3", order.Status, order.IsPreOrder, order.Onumber)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("move order error, can't move preoreder to order, %w", err)
	}

	_, err = tx.ExecContext(ctx, "UPDATE bonuses SET bonus_accrual = $1 WHERE onumber = $2", order.Bonus.Accrual, order.Onumber)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("move order error, can't set bonuses for this order, %w", err)
	}

	tx.Commit()

	return

}
