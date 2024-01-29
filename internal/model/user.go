package model

import (
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

type User struct {
	UUID        *uuid.UUID      `json:"-" db:"user_id"`
	Login       string          `json:"login" db:"login"`
	Password    string          `json:"password"`
	PassHash    string          `db:"password_hash"`
	Withdrawals decimal.Decimal `db:"withdrawals"`
	Bonuses     decimal.Decimal `db:"bonuses"`
}
