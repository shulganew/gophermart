package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"github.com/shulganew/gophermart/internal/model"
	"go.uber.org/zap"
)

func (base *Repo) AddUser(ctx context.Context, login string, hash string) (*uuid.UUID, error) {
	query := `
	INSERT INTO users (login, password_hash) 
	VALUES ($1, $2)
	RETURNING user_id
	`
	userID := &uuid.UUID{}

	err := base.master.GetContext(ctx, userID, query, login, hash)
	if err != nil {
		var pgErr *pq.Error
		// if URL exist in DataBase
		if errors.As(err, &pgErr) && pgerrcode.UniqueViolation == pgErr.Code {
			return nil, pgErr
		}
		return nil, fmt.Errorf("error adding user to Storage: %w", err)
	}
	return userID, nil
}

// Retrive User by login.
func (base *Repo) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	query := `
	SELECT user_id, password_hash 
	FROM users 
	WHERE login = $1
	`
	user := model.User{Login: login}
	zap.S().Infoln("user login: ", login)
	err := base.master.GetContext(ctx, &user, query, login)
	if err != nil {
		return nil, fmt.Errorf("error during get user by login from storage. User not valid: %w", err)
	}
	return &user, nil
}


