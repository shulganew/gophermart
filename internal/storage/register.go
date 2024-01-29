package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"github.com/shulganew/gophermart/internal/model"
	"go.uber.org/zap"
)

func (base *Repo) AddUser(ctx context.Context, user model.User, hash string) error {
	_, err := base.master.ExecContext(ctx, "INSERT INTO users (user_id, login, password_hash) VALUES ($1, $2, $3)", user.UUID, user.Login, hash)
	if err != nil {
		var pgErr *pq.Error
		// if URL exist in DataBase
		if errors.As(err, &pgErr) && pgerrcode.UniqueViolation == pgErr.Code {
			return pgErr
		}
		return fmt.Errorf("error adding user to Storage: %w", err)
	}
	return nil
}

// Retrive User by login
func (base *Repo) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	user := model.User{Login: login}
	zap.S().Infoln("user login: ", login)
	err := base.master.MustBegin().GetContext(ctx, &user, "SELECT user_id, password_hash FROM users WHERE login = $1", login)
	if err != nil {
		return nil, fmt.Errorf("error during get user by login from storage. User not valid: %w", err)
	}
	return &user, nil
}
