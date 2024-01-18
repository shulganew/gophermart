package storage

import (
	"context"
	"errors"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgerrcode"
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

	return
}

func (base *RepoMarket) Start(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	err := base.master.Ping(ctx)
	defer cancel()
	return err
}

func (base *RepoMarket) GetOrder() {

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

func (base *RepoMarket) SetOrder(ctx context.Context, userID *uuid.UUID, order string) error {

	_, err := base.master.Exec(ctx, "INSERT INTO orders (user_id, onumber, uploaded) VALUES ($1, $2, $3)", userID, order, time.Now())
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

	_, err = base.master.Exec(ctx, "INSERT INTO bonuses (onumber) VALUES ($1)", order)
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

// Load all orders with not finished preparation status
func (base *RepoMarket) LoadOrders(ctx context.Context) ([]model.Order, error) {

	query := `
	SELECT users.user_id, orders.onumber, orders.uploaded, orders.status
	FROM orders INNER JOIN users ON orders.user_id = users.user_id 
	WHERE status = 'NEW' OR status = 'REGISTERED' OR status = 'PROCESSING'
	`

	rows, err := base.master.Query(ctx, query)

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

func (base *RepoMarket) UpdateOrderStatus(ctx context.Context, order *model.Order, status *model.Status, accural *decimal.Decimal) (err error) {
	_, err = base.master.Exec(ctx, "UPDATE orders SET status = $1 WHERE onumber = $2", status.String(), order.Onumber)
	if err != nil {
		zap.S().Errorln("UPDATE order Status error: ", err)
		return err
	}

	if accural != nil {
		_, err = base.master.Exec(ctx, "UPDATE bonuses SET  bonus_accural = $1 WHERE onumber = $2", accural, order.Onumber)
		if err != nil {
			zap.S().Errorln("UPDATE order Status error: ", err)
			return err
		}
	}

	return nil
}
