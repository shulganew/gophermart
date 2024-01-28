package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shulganew/gophermart/internal/model"
)

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
