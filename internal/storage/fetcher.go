package storage

import (
	"context"
	"fmt"

	"github.com/shulganew/gophermart/internal/model"
)

// Load all orders with not finished preparation status.
func (base *Repo) LoadPocessing(ctx context.Context) ([]model.Order, error) {

	orders := make([]model.Order, 0)
	query := `
	SELECT  user_id, order_number
		FROM orders 
		WHERE (status = 'NEW' OR status = 'REGISTERED' OR status = 'PROCESSING') AND is_preorder = FALSE
	`
	err := base.master.SelectContext(ctx, &orders, query)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (base *Repo) UpdateStatus(ctx context.Context, order string, status model.Status) (err error) {

	_, err = base.master.ExecContext(ctx, "UPDATE orders SET status = $1 WHERE order_number = $2", status, order)
	if err != nil {
		return fmt.Errorf("can't update orders status, %w", err)
	}

	return
}
