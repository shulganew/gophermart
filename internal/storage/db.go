package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgerrcode"
	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
	"github.com/shulganew/gophermart/internal/model"
	"go.uber.org/zap"
)

type RepoMarket struct {
	master *pgx.Conn
}

func NewRepo(ctx context.Context, master *pgx.Conn) (*RepoMarket, error) {
	db := RepoMarket{master: master}
	err := db.Start(ctx)
	return &db, err
}

// Init Database
func InitDB(ctx context.Context, dsn string) (db *pgx.Conn, err error) {

	db, err = pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	// Register Numeric and decimal support
	pgxdecimal.Register(db.TypeMap())

	// Create tables for Market if not exist

	query := `
	DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'processing') THEN
			CREATE TYPE processing AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED', 'REGISTERED');
		END IF;
	END$$
	`
	_, err = db.Exec(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error create processing enum:  %w", err)
	}

	query = `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL, 
		user_id UUID NOT NULL UNIQUE, 
		login TEXT NOT NULL UNIQUE, 
		password TEXT NOT NULL
		);
		`

	_, err = db.Exec(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error create users %w", err)
	}

	query = `
	CREATE TABLE IF NOT EXISTS orders (
		id SERIAL, 
		user_id UUID NOT NULL REFERENCES users(user_id),
		onumber VARCHAR(20) NOT NULL UNIQUE,
		isPreorder BOOLEAN NOT NULL, 
		uploaded TIMESTAMPTZ NOT NULL,
		status processing NOT NULL DEFAULT 'NEW'
		);
		`
	_, err = db.Exec(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error create orders %w", err)
	}

	query = `
	CREATE TABLE IF NOT EXISTS bonuses (
		id SERIAL, 
		onumber VARCHAR(20) NOT NULL REFERENCES orders(onumber) UNIQUE,
		bonus_used NUMERIC DEFAULT 0,
		bonus_accrual NUMERIC DEFAULT 0
	);
	`
	_, err = db.Exec(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error create bonuses %w", err)
	}

	return
}

func (base *RepoMarket) Start(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	err := base.master.Ping(ctx)
	defer cancel()
	return err
}

func (base *RepoMarket) Register(ctx context.Context, user model.User) error {

	_, err := base.master.Exec(ctx, "INSERT INTO users (user_id, login, password) VALUES ($1, $2, $3)", user.UUID, user.Login, user.Password)
	if err != nil {

		var pgErr *pgconn.PgError

		// if URL exist in DataBase
		if errors.As(err, &pgErr) && pgerrcode.UniqueViolation == pgErr.Code {
			return pgErr
		}

		zap.S().Errorln("Insert user error: ", err)
		return err
	}
	return nil
}

// Retrive User by login

func (base *RepoMarket) GetByLogin(ctx context.Context, login string) (*model.User, error) {

	row := base.master.QueryRow(ctx, "SELECT user_id, password FROM users WHERE login = $1", login)
	user := model.User{Login: login}
	err := row.Scan(&user.UUID, &user.Password)
	if err != nil {
		zap.S().Errorln("User not valid: ", err)
		return nil, err
	}

	return &user, nil
}

func (base *RepoMarket) GetOrders(ctx context.Context, userID *uuid.UUID) ([]model.Order, error) {
	query := `
	SELECT users.user_id, orders.onumber, orders.uploaded, orders.status, bonuses.bonus_used, bonuses.bonus_accrual
		FROM orders 
		INNER JOIN users ON orders.user_id = users.user_id
		INNER JOIN bonuses ON orders.onumber = bonuses.onumber
		WHERE isPreorder = FALSE AND orders.user_id = $1
		ORDER BY orders.uploaded DESC
		`

	rows, err := base.master.Query(ctx, query, userID)
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
		order.Status.SetStatus(status)
		order.Bonus = model.NewBonus(&used, &accrual)
		orders = append(orders, order)
	}

	return orders, nil

}
func (base *RepoMarket) SetOrder(ctx context.Context, userID *uuid.UUID, order string, isPreOrder bool, withdrawn *decimal.Decimal) error {

	_, err := base.master.Exec(ctx, "INSERT INTO orders (user_id, onumber, isPreorder, uploaded) VALUES ($1, $2, $3, $4)", userID, order, isPreOrder, time.Now())
	if err != nil {
		var pgErr *pgconn.PgError
		errors.As(err, &pgErr)
		// if order exist in DataBase
		if errors.As(err, &pgErr) && pgerrcode.UniqueViolation == pgErr.Code {
			return pgErr
		}

		zap.S().Errorln("Insert order error: ", err)
		return err
	}

	_, err = base.master.Exec(ctx, "INSERT INTO bonuses (onumber, bonus_used) VALUES ($1, $2)", order, *withdrawn)
	if err != nil {
		zap.S().Errorln("Insert order error: ", err)
		return err
	}

	return nil
}

func (base *RepoMarket) IsExistForUser(ctx context.Context, userID *uuid.UUID, order string) (isExist bool, err error) {

	var ordersn int
	row := base.master.QueryRow(ctx, "SELECT count(*) FROM orders WHERE user_id = $1 AND onumber = $2", userID, order)
	err = row.Scan(&ordersn)
	if err != nil {
		zap.S().Errorln("Error during order search for user: ", userID, order, err)
		return true, err
	}
	if ordersn == 0 {
		return false, nil
	}

	return true, nil
}

func (base *RepoMarket) IsExistForOtherUsers(ctx context.Context, userID *uuid.UUID, order string) (isExist bool, err error) {
	var ordersn int
	row := base.master.QueryRow(ctx, "SELECT count(*) FROM orders WHERE user_id != $1 AND onumber = $2", userID, order)
	err = row.Scan(&ordersn)
	if err != nil {
		zap.S().Errorln("Error during order search for user: ", userID, order, err)
		return true, err
	}
	if ordersn == 0 {
		return false, nil
	}

	return true, nil
}

func (base *RepoMarket) IsPreOrder(ctx context.Context, userID *uuid.UUID, order string) (isPreOrder bool, err error) {
	query := `
	SELECT count(orders.onumber)
	FROM orders INNER JOIN users ON orders.user_id = users.user_id 
	WHERE users.user_id = $1 AND orders.onumber = $2 AND isPreorder = TRUE
	`
	row := base.master.QueryRow(ctx, query, userID, order)
	var n int
	err = row.Scan(&n)

	return n == 1, err
}

func (base *RepoMarket) MovePreOrder(ctx context.Context, order *model.Order) (err error) {

	_, err = base.master.Exec(ctx, "UPDATE orders SET status = $1, isPreorder = $2 WHERE onumber = $3", order.Status.String(), order.IsPreOrder, order.Onumber)
	if err != nil {
		zap.S().Errorln("UPDATE preoreder error: ", err)
		return err
	}

	_, err = base.master.Exec(ctx, "UPDATE bonuses SET bonus_accrual = $1 WHERE onumber = $2", order.Bonus.Accrual, order.Onumber)
	if err != nil {
		zap.S().Errorln("UPDATE preoreder's bonus error: ", err)
		return err
	}

	return

}

// Load all orders with not finished preparation status.
func (base *RepoMarket) LoadPocessing(ctx context.Context) ([]model.Order, error) {

	query := `
	SELECT users.user_id, orders.onumber, orders.uploaded, orders.status
	FROM orders INNER JOIN users ON orders.user_id = users.user_id 
	WHERE (status = 'NEW' OR status = 'REGISTERED' OR status = 'PROCESSING') AND isPreorder = FALSE
	`

	rows, err := base.master.Query(ctx, query)
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
		order.Status.SetStatus(status)
		orders = append(orders, order)
	}

	return orders, nil
}

func (base *RepoMarket) UpdateStatus(ctx context.Context, order *model.Order, accrual *decimal.Decimal) (err error) {

	_, err = base.master.Exec(ctx, "UPDATE orders SET status = $1 WHERE onumber = $2", order.Status.String(), order.Onumber)
	if err != nil {
		zap.S().Errorln("UPDATE order Status error: ", err)
		return err
	}

	if accrual != nil {
		_, err = base.master.Exec(ctx, "UPDATE bonuses SET  bonus_accrual = $1 WHERE onumber = $2", accrual, order.Onumber)
		if err != nil {
			zap.S().Errorln("UPDATE order Status error: ", err)
			return err
		}
	}

	return nil
}

func (base *RepoMarket) GetAccruals(ctx context.Context, userID *uuid.UUID) (accrual *decimal.Decimal, err error) {
	query := `
	SELECT SUM(bonuses.bonus_accrual)
		FROM orders 
		INNER JOIN users ON orders.user_id = users.user_id
		INNER JOIN bonuses ON orders.onumber = bonuses.onumber;
	`

	row := base.master.QueryRow(ctx, query)

	err = row.Scan(&accrual)
	if err != nil {
		return nil, err
	}

	// If Postgres SIM is 0, it return nil
	if accrual == nil {
		return &decimal.Zero, nil
	}

	return accrual, nil
}

func (base *RepoMarket) GetWithdrawns(ctx context.Context, userID *uuid.UUID) (withdrawn *decimal.Decimal, err error) {
	query := `
	SELECT SUM(bonuses.bonus_used)
		FROM orders 
		INNER JOIN users ON orders.user_id = users.user_id
		INNER JOIN bonuses ON orders.onumber = bonuses.onumber;
	`

	row := base.master.QueryRow(ctx, query)

	err = row.Scan(&withdrawn)
	if err != nil {
		return nil, err
	}

	return withdrawn, nil
}

func (base *RepoMarket) Withdrawals(ctx context.Context, userID *uuid.UUID) (wds []model.Withdrawals, err error) {

	query := `
	SELECT  orders.onumber, bonuses.bonus_used, orders.uploaded
		FROM orders 
		INNER JOIN users ON orders.user_id = users.user_id
		INNER JOIN bonuses ON orders.onumber = bonuses.onumber
		ORDER BY orders.uploaded DESC
		`

	rows, err := base.master.Query(ctx, query)

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

	return
}
