package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/shopspring/decimal"
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/model"

	"go.uber.org/zap"
)

type Repo struct {
	master *sqlx.DB
}

func NewRepo(ctx context.Context, master *sqlx.DB) (*Repo, error) {
	db := Repo{master: master}
	err := db.Start(ctx)
	return &db, err
}

// Init Database
func InitDB(ctx context.Context, dsn string, migrationdns string) (db *sqlx.DB, err error) {

	//Init connection for admin user for prepare databse and make migrations
	initdb, err := goose.OpenDBWithDriver(config.DataBaseType, migrationdns)
	if err != nil {
		zap.S().Fatalln("goose: failed to open DB: %v\n", err)
	}

	defer func() {
		if err := initdb.Close(); err != nil {
			zap.S().Fatalln("goose: failed to close DB: %v\n", err)
		}
	}()

	//Init database migrations
	if err := goose.UpContext(ctx, initdb, "migrations"); err != nil { //
		zap.S().Fatalln("Error make databes migrations before starting Market app: ", err)
	} else {
		zap.S().Infoln("Migrations update...")
	}

	//Connection for Gophermart
	db, err = sqlx.Connect(config.DataBaseType, dsn)
	if err != nil {
		return nil, err
	}

	// Create tables for Market if not exist
	query := `
	DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'processing') THEN
			CREATE TYPE processing AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED', 'REGISTERED');
		END IF;
	END$$
	`
	_, err = db.ExecContext(ctx, query)
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

	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error create users %w", err)
	}

	query = `
	CREATE TABLE IF NOT EXISTS orders (
		id SERIAL, 
		user_id UUID NOT NULL REFERENCES users(user_id),
		onumber VARCHAR(20) NOT NULL UNIQUE,
		is_preorder BOOLEAN NOT NULL, 
		uploaded TIMESTAMPTZ NOT NULL,
		status processing NOT NULL DEFAULT 'NEW'
		);
		`
	_, err = db.ExecContext(ctx, query)
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
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error create bonuses %w", err)
	}

	return
}

func (base *Repo) Start(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	err := base.master.Ping()
	defer cancel()
	return err
}

func (base *Repo) AddUser(ctx context.Context, user model.User) error {

	_, err := base.master.ExecContext(ctx, "INSERT INTO users (user_id, login, password) VALUES ($1, $2, $3)", user.UUID, user.Login, user.Password)
	if err != nil {

		var pgErr *pgconn.PgError

		// if URL exist in DataBase
		if errors.As(err, &pgErr) && pgerrcode.UniqueViolation == pgErr.Code {
			return pgErr
		}

		return fmt.Errorf("Error adding user to Storage: %w", err)
	}
	return nil
}

// Retrive User by login
func (base *Repo) GetByLogin(ctx context.Context, login string) (*model.User, error) {

	row := base.master.QueryRowContext(ctx, "SELECT user_id, password FROM users WHERE login = $1", login)
	user := model.User{Login: login}
	err := row.Scan(&user.UUID, &user.Password)
	if err != nil {
		return nil, fmt.Errorf("Error during get user by login from storage. User not valid: %w", err)
	}

	return &user, nil
}

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

	return orders, nil

}
func (base *Repo) AddOrder(ctx context.Context, userID *uuid.UUID, order string, isPreOrder bool, withdrawn *decimal.Decimal) error {

	_, err := base.master.ExecContext(ctx, "INSERT INTO orders (user_id, onumber, is_preorder, uploaded) VALUES ($1, $2, $3, $4)", userID, order, isPreOrder, time.Now())
	if err != nil {
		var pgErr *pgconn.PgError
		errors.As(err, &pgErr)
		// if order exist in DataBase
		if errors.As(err, &pgErr) && pgerrcode.UniqueViolation == pgErr.Code {
			return pgErr
		}
		return fmt.Errorf("Error during set order to Storage, error: %w", err)
	}

	_, err = base.master.ExecContext(ctx, "INSERT INTO bonuses (onumber, bonus_used) VALUES ($1, $2)", order, *withdrawn)
	if err != nil {
		return fmt.Errorf("Error during add bonuses order to Storage, error: %w", err)
	}

	return nil
}

func (base *Repo) IsExistForUser(ctx context.Context, userID *uuid.UUID, order string) (isExist bool, err error) {

	var ordersn int
	row := base.master.QueryRowContext(ctx, "SELECT count(*) FROM orders WHERE user_id = $1 AND onumber = $2", userID, order)
	err = row.Scan(&ordersn)
	if err != nil {
		return true, fmt.Errorf("Error during order search for user: %w", err)
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
		return true, fmt.Errorf("Error during order search for user: %w", err)
	}
	if ordersn == 0 {
		return false, nil
	}

	return true, nil
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
		return fmt.Errorf("Move order error, begin transaction error, %w", err)
	}

	_, err = tx.ExecContext(ctx, "UPDATE orders SET status = $1, is_preorder = $2 WHERE onumber = $3", order.Status, order.IsPreOrder, order.Onumber)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Move order error, can't move preoreder to order, %w", err)
	}

	_, err = tx.ExecContext(ctx, "UPDATE bonuses SET bonus_accrual = $1 WHERE onumber = $2", order.Bonus.Accrual, order.Onumber)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Move order error, can't set bonuses for this order, %w", err)
	}

	tx.Commit()

	return

}

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

	return orders, nil
}

func (base *Repo) UpdateStatus(ctx context.Context, order *model.Order, accrual *decimal.Decimal) (err error) {

	tx, err := base.master.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Can't update orders status, begin transaction error, %w", err)
	}

	_, err = tx.ExecContext(ctx, "UPDATE orders SET status = $1 WHERE onumber = $2", order.Status, order.Onumber)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Can't update orders status, %w", err)
	}

	if accrual != nil {
		_, err = tx.ExecContext(ctx, "UPDATE bonuses SET bonus_accrual = $1 WHERE onumber = $2", accrual, order.Onumber)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("Can't update bonuses during update status %w", err)
		}
	}

	tx.Commit()

	return nil
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
